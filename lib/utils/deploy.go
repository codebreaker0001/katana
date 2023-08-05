package utils

import (
	"context"
	"log"

	g "github.com/sdslabs/katana/configs"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DeployChallenge(challengeName, teamName string, firstPatch bool, replicas int32) error {

	teamNamespace := teamName + "-ns"
	kubeclient, err := GetKubeClient()
	if err != nil {
		return err
	}

	deploymentsClient := kubeclient.AppsV1().Deployments(teamNamespace)
	imageName := g.HarborConfig.Hostname + "/katana/" + challengeName + ":latest"
	if firstPatch {
		/// Retrieve the existing deployment
		existingDeployment, err := deploymentsClient.Get(context.TODO(), challengeName, metav1.GetOptions{})
		if err != nil {
			log.Println("Error in retrieving existing deployment.")
			log.Println(err)
			return err
		}

		existingDeployment.Spec.Template.Spec.Containers[0].Image = teamName + "/" + challengeName

		_, err = deploymentsClient.Update(context.TODO(), existingDeployment, metav1.UpdateOptions{})
		if err != nil {
			log.Println("Error in updating deployment.")
			log.Println(err)
			return err
		}

		log.Println("Updated deployment with new image.")
		return nil
	}

	manifest := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: teamNamespace,
			Name:      challengeName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": challengeName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": challengeName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            challengeName + "-" + teamName,
							Image:           imageName,
							ImagePullPolicy: v1.PullPolicy("IfNotPresent"),
							Ports: []v1.ContainerPort{
								{
									Name:          "challenge-port",
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	deployment, err := deploymentsClient.Get(context.TODO(), challengeName, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
		return err
	}
	if deployment.Name == challengeName {
		log.Println("Deployment already exists for the challenge " + challengeName + " in namespace " + teamNamespace)
		return nil
	}

	// Create Deployment
	log.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), manifest, metav1.CreateOptions{})

	if err != nil {
		log.Println("Unable to create deployement")
		panic(err)
	}

	log.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName()+" in namespace "+teamNamespace)
	return nil

}
