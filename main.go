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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	JobStatusQueued     = "queued"
	JobStatusProcessing = "processing"
	JobStatusFinished   = "finished"
	JobStatusError      = "error"
)

// Job - stores all job data, will be stored in redis
type Job struct {
	Input     string `json:"input"`
	Status    string `json:"status"`
	StartTime int64  `json:"startTime"`
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

	// TODO start model
	createPod()

	json.NewEncoder(w).Encode(map[string]string{"id": jobId})
}

// curl -X POST http://localhost:10000/push \
//    -H 'Content-Type: application/json' \
//    -d '{"input": "whatever" }'

func redisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
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

	// TODO check if job complete
	latency := time.Now().Unix() - job.StartTime

	json.NewEncoder(w).Encode(map[string]string{"input": job.Input, "latency": strconv.Itoa(int(latency))})
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

	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {
	handleRequests()
}

// TODO need to figure out how to move all this into correct directory structure

func createPod() {
	kubeconfig := "/Users/britonns/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	pod := getPodObject()

	// now create the pod in kubernetes cluster using the clientset
	pod, err = clientset.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println(pod)
}

func getPodObject() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test-pod",
			Namespace: "default",
			Labels: map[string]string{
				"app": "demo",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "busybox",
					Image:           "busybox",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command: []string{
						"sleep",
						"3600",
					},
				},
			},
		},
	}
}
