package generate

import (
	"fmt"
	"testing"
)

func TestClient_SendMessage(t *testing.T) {
	client := New(PlatformGemini, "AIzaSyBZQFbZcecU145agAvqVe0rlTSj8ZhExZw", "", "http://127.0.0.1:7890", "")
	message, err := client.SendMessage("test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	fmt.Println(message)
}

// func TestClient_sendGeminiMessage(t *testing.T) {
// 	client := New(ClientTypeGemini, "test", "", "")
// 	_, err := client.sendGeminiMessage("test")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
// }

// func TestClient_sendOpenAIMessage(t *testing.T) {
// 	client := New(ClientTypeOpenAI, "test", "", "")
// 	_, err := client.sendOpenAIMessage("test")
// 	if err != nil {
// 		t.Errorf("unexpected error: %v", err)
// 	}
// }
