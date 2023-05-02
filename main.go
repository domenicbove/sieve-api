// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	JobStatusQueued = "queued"
	JobStatusError  = "error"
)

// Job - stores all job data, will be stored in redis
type Job struct {
	Input     string `json:"input"`
	Output    string `json:"output"`
	Status    string `json:"status"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}

func pushNewJob(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var job Job
	json.Unmarshal(reqBody, &job)
	job.Status = JobStatusQueued
	job.StartTime = time.Now().Unix()
	jobId := uuid.New().String()

	fmt.Println("recieved push job request")

	jsonJob, err := json.Marshal(job)
	if err != nil {
		fmt.Println(err)
	}

	// add Job Details to Redis
	client := redisClient()
	err = client.Set(jobId, jsonJob, 0).Err()
	if err != nil {
		fmt.Println(err)
	}

	// add jobId to jobs list
	client.RPush("jobs", jobId)

	// create job based off of queue length and number of setup models
	jobQueueLength := int(client.LLen("jobs").Val())
	fmt.Printf("number of jobs in the queue: %d\n", jobQueueLength)

	createModelJob(jobId, job.Input, jobQueueLength)

	json.NewEncoder(w).Encode(map[string]string{"id": jobId})
}

func redisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
}

func getJob(jobId string) (Job, error) {

	var job Job

	client := redisClient()
	val, err := client.Get(jobId).Result()
	if err != nil {
		return job, err
	}

	json.Unmarshal([]byte(val), &job)

	return job, nil
}

func jobData(w http.ResponseWriter, r *http.Request) {
	fmt.Println("recieved job data request")

	vars := mux.Vars(r)

	job, err := getJob(vars["id"])
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		return
	}

	latency := job.EndTime - job.StartTime

	json.NewEncoder(w).Encode(map[string]string{
		"input":   job.Input,
		"latency": strconv.Itoa(int(latency)),
		"output":  job.Output,
	})
}

func jobStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Println("recieved job status request")

	vars := mux.Vars(r)

	job, err := getJob(vars["id"])
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": job.Status})
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/push", pushNewJob).Methods("POST")
	myRouter.HandleFunc("/data/{id}", jobData)
	myRouter.HandleFunc("/status/{id}", jobStatus)

	myRouter.HandleFunc("/health", healthHandler)
	myRouter.HandleFunc("/readiness", readinessHandler)

	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	handleRequests()
}

func createModelJob(jobId, jobInput string, jobQueueLength int) {

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}

	listOptions := metav1.ListOptions{
		LabelSelector: "setup-complete=true",
	}
	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), listOptions)
	if err != nil {
		fmt.Println(err)
	}

	numSetupPods := len(pods.Items)
	fmt.Printf("number of setup models: %d\n", numSetupPods)

	if numSetupPods > 0 {
		if jobQueueLength*10/numSetupPods < 52 {
			return
		}
	}

	modelJob := getModelJob(jobId, jobInput)

	_, err = clientset.BatchV1().Jobs(modelJob.Namespace).Create(context.TODO(), modelJob, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
	}

}

func getModelJob(jobId, jobInput string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("model-%s", jobId[:5]),
			Namespace: "default",
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "model",
							Image:           "model:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"python"},
							Args: []string{
								"src/model.py",
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
}
