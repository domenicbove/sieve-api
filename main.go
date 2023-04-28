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

	jsonJob, err := json.Marshal(job)
	if err != nil {
		fmt.Println(err)
	}

	client := redisClient()
	err = client.Set(jobId, jsonJob, 0).Err()
	if err != nil {
		fmt.Println(err)
	}

	createPod(jobId, job.Input)

	json.NewEncoder(w).Encode(map[string]string{"id": jobId})
}

// curl -X POST http://localhost:8080/push \
//    -H 'Content-Type: application/json' \
//    -d '{"input": "whatever" }'

func redisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		// TODO use env vars here, switch back to localhost for local dev
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
	vars := mux.Vars(r)

	job, err := getJob(vars["id"])
	if err != nil {
		fmt.Println(err)
		// TODO more intelligent error handling
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
	vars := mux.Vars(r)

	job, err := getJob(vars["id"])
	if err != nil {
		fmt.Println(err)
		// TODO more intelligent error handling
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

// TODO need to figure out how to move all this into correct directory structure

func createPod(jobId, jobInput string) {

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println(err)
	}

	// running locally...
	// kubeconfig := "/Users/britonns/.kube/config"
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
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
								jobId,
								jobInput,
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
}
