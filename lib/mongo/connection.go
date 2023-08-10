package mongo

import (
	"context"
	"log"
	"time"

	"github.com/sdslabs/katana/configs"
	"github.com/sdslabs/katana/lib/utils"
	"github.com/sdslabs/katana/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx, _ = context.WithTimeout(context.Background(), time.Duration(configs.KatanaConfig.TimeOut)*time.Second)
var client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+configs.MongoConfig.Username+":"+configs.MongoConfig.Password+"@"+utils.GetKatanaLoadbalancer()+":27017/?directConnection=true&appName=mongosh+"+configs.MongoConfig.Version))
var link = client.Database(projectDatabase)

func setupAdmin() {
	adminUser := configs.AdminConfig
	pwd, err := utils.HashPassword(adminUser.Password)
	if err != nil {
		log.Fatal(err)
	}

	admin := types.AdminUser{
		Username: adminUser.Username,
		Password: pwd,
	}

	if _, err = AddAdmin(context.Background(), admin); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("admin privileges have been given to username: %s", admin.Username)
	}
}

func setup(LoadbalancerIP string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configs.KatanaConfig.TimeOut)*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Println("MongoDB connection was not established")
		log.Println("Error: ", err)
		time.Sleep(time.Duration(configs.KatanaConfig.TimeOut) * time.Second)
		setup(LoadbalancerIP)
	} else {
		log.Println("MongoDB Connection Established")
		setupAdmin()
	}
}

func Test() {
	collection := link.Collection(TeamsCollection)
	res, err := collection.InsertOne(context.Background(), bson.M{"team": "ctfteam-0", "password": "passloll"})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)
}

func Init(LoadbalancerIP string) {
	go setup(LoadbalancerIP)
}
