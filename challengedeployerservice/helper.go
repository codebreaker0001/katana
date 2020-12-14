package challengedeployerservice

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getClient(pathToCfg string) (*kubernetes.Clientset, error) {
	if pathToCfg == "" {
		pathToCfg = filepath.Join(
			os.Getenv("HOME"), ".kube", "config",
		)
	}
	config, err := clientcmd.BuildConfigFromFlags("", pathToCfg)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func getPods(lbls map[string]string) ([]v1.Pod, error) {
	set := labels.Set(lbls)
	pods, err := kubeclient.CoreV1().Pods(config.KubeNameSpace).List(context.Background(), metav1.ListOptions{LabelSelector: set.AsSelector().String()})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// Clone the provided remote repository into repos/<local>
func clone(remote string, local string, auth *githttp.BasicAuth) error {
	cloneConfig := &git.CloneOptions{
		URL:      remote,
		Progress: os.Stdout,
		Auth:     auth,
	}
	tmpdir, err := ioutil.TempDir("tmp", local)
	if err != nil {
		return err
	}
	fmt.Printf("Temporary directory created: %s", tmpdir)
	if _, err := git.PlainClone(fmt.Sprintf(tmpdir), false, cloneConfig); err != nil {
		return err
	}
	return compressAndMove(tmpdir, fmt.Sprintf("challenges/%s.zip", local))
}

// Compress the src directory into <dst>.zip
func compressAndMove(src string, dst string) error {
	outfile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer outfile.Close()

	w := zip.NewWriter(outfile)
	defer w.Close()

	n := len(src)
	if src[n-1] == '/' {
		src = src[:n-1]
		n -= 1
	}

	addToZip := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// TODO: Check for arbitrary read/ write

		f, err := w.Create(path[n:])
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}

	if err = filepath.Walk(src, addToZip); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func sendFile(file *os.File, params map[string]string, filename, uri string) error {
	client := &http.Client{}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(filename, filename)

	if err != nil {
		return err
	}

	if _, err = io.Copy(part, file); err != nil {
		return err
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	if err = writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client.Do(req)
	return nil
}

func broadcast(file string) error {
	chal, err := os.Open(filepath.Join("challenges", file))
	if err != nil {
		return err
	}
	defer chal.Close()

	lbl := make(map[string]string)
	lbl["app"] = config.TeamLabel
	teamPods, err := getPods(lbl)
	if err != nil {
		return err
	}

	addresses := []string{}
	for _, pod := range teamPods {
		addresses = append(addresses, fmt.Sprintf("%s:%d", pod.Status.PodIP, config.TeamClientPort))
	}

	addressesEncoded, err := json.Marshal(addresses)
	if err != nil {
		return err
	}

	params := make(map[string]string)
	params["targets"] = string(addressesEncoded)

	return sendFile(chal, params, file, fmt.Sprintf("%s:%d", config.KubeHost, config.BroadcastPort))
}
