package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/gorilla/mux"
)

type SubscriptionData struct {
	Subscription webpush.Subscription `json:"subscription"`
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
	payload := []byte(fmt.Sprintf("Hello from the Go server! It's %s", time.Now().Format(time.RFC822)))

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount int

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
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			successCount++
			mu.Unlock()
		}(sub)
		return true
	})

	w.WriteHeader(http.StatusOK)
	go func() {
		wg.Wait()
		fmt.Printf("Notification sent successfully to %d subscriptions.\n", successCount)
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
