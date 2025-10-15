package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/gorilla/mux"
)

type SubscriptionData struct {
	Subscription webpush.Subscription `json:"subscription"`
}

type NotificationRequest struct {
	Message string `json:"message"`
}

type VapidKeys struct {
	PrivateKey string
	PublicKey  string
}

var (
	subscriptions sync.Map
	vapidKeys     *VapidKeys
)

func init() {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatal("Failed to generate VAPID keys:", err)
	}
	vapidKeys = &VapidKeys{PrivateKey: privateKey, PublicKey: publicKey}
	fmt.Println("VAPID Public Key:", vapidKeys.PublicKey)
}

func getVAPIDPublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"publicKey": vapidKeys.PublicKey})
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	var data SubscriptionData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	subscriptions.Store(data.Subscription.Endpoint, data.Subscription)
	fmt.Printf("Subscription with endpoint %s saved successfully.\n", data.Subscription.Endpoint)

	w.WriteHeader(http.StatusOK)
}

func notifyHandler(w http.ResponseWriter, r *http.Request) {
	var req NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payload := []byte(req.Message)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount int
	var failureCount int

	subscriptions.Range(func(key, value interface{}) bool {
		sub := value.(webpush.Subscription)
		wg.Add(1)
		go func(s webpush.Subscription) {
			defer wg.Done()
			resp, err := webpush.SendNotification(payload, &s, &webpush.Options{
				Subscriber:      "example@example.com",
				VAPIDPublicKey:  vapidKeys.PublicKey,
				VAPIDPrivateKey: vapidKeys.PrivateKey,
				TTL:             30,
			})
			if err != nil {
				log.Printf("Failed to send notification to endpoint %s: %s", s.Endpoint, err)
				mu.Lock()
				failureCount++
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= http.StatusBadRequest {
				log.Printf("Failed to send notification to endpoint %s: received error status code %d", s.Endpoint, resp.StatusCode)
				mu.Lock()
				failureCount++
				mu.Unlock()
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}(sub)
		return true
	})

	w.WriteHeader(http.StatusOK)
	go func() {
		wg.Wait()
		fmt.Printf("Notification sent successfully to %d subscriptions, %d failures.\n", successCount, failureCount)
	}()
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/vapid-public-key", getVAPIDPublicKeyHandler).Methods("GET")
	router.HandleFunc("/subscribe", subscribeHandler).Methods("POST")
	router.HandleFunc("/notify-all", notifyHandler).Methods("POST")

	fmt.Println("Server listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
