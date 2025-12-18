package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

var authToken string
var createdDiaryID uint

func main() {
	fmt.Println("开始 API 接口测试...")
	fmt.Println("--------------------------------------------------")

	// 1. 注册
	register()

	// 2. 登录
	login()

	// 3. 创建标签
	createTag("测试标签")

	// 4. 创建待办事项
	createTodo("完成API测试", "编写并运行测试脚本")

	// 5. 创建日记
	createDiary("今天天气不错", "心情很好，代码一次过！", "测试标签")

	// 6. 获取日记列表
	listDiaries()

	// 7. 搜索日记
	searchDiary("今天")

	// 8. 获取统计看板
	getDashboardStats()

	// 9. 获取日记详情
	if createdDiaryID != 0 {
		getDiary(createdDiaryID)

		// 10. 更新日记
		updateDiary(createdDiaryID, "今天天气真棒", "心情非常好，代码测试通过！")

		// 11. 删除日记
		deleteDiary(createdDiaryID)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("测试完成！")
}

func register() {
	fmt.Println("[1/9] 测试注册接口...")
	username := fmt.Sprintf("testuser_%d", time.Now().Unix())
	body := map[string]string{
		"username": username,
		"password": "password123",
	}
	resp := request("POST", "/api/register", nil, body)
	printResponse(resp)
}

func login() {
	fmt.Println("[2/9] 测试登录接口...")
	fixedUser := "apitest_user"
	fixedPass := "123456"

	// 尝试注册固定账号
	request("POST", "/api/register", nil, map[string]string{
		"username": fixedUser,
		"password": fixedPass,
	})

	// 登录
	body := map[string]string{
		"username": fixedUser,
		"password": fixedPass,
	}
	resp := request("POST", "/api/login", nil, body)

	if resp != nil {
		if token, ok := resp["data"].(map[string]interface{})["token"].(string); ok {
			authToken = token
			fmt.Println("✅ 登录成功，Token获取成功")
		} else {
			fmt.Println("❌ 登录失败或未获取到Token")
		}
	}
}

func createTag(name string) {
	fmt.Println("[3/9] 测试创建标签...")
	body := map[string]string{
		"name": name,
	}
	request("POST", "/api/tags", &authToken, body)
}

func createTodo(title, desc string) {
	fmt.Println("[4/9] 测试创建待办...")
	body := map[string]interface{}{
		"title":       title,
		"description": desc,
		"due_date":    time.Now().Add(24 * time.Hour),
	}
	request("POST", "/api/todos", &authToken, body)
}

func createDiary(title, content, tagName string) {
	fmt.Println("[5/9] 测试创建日记...")
	body := map[string]interface{}{
		"title":     title,
		"content":   content,
		"weather":   "Sunny",
		"mood":      "Happy",
		"location":  "Home",
		"date":      time.Now(),
		"is_public": true,
		"tags":      []string{tagName},
	}
	resp := request("POST", "/api/diaries", &authToken, body)
	if resp != nil {
		if data, ok := resp["data"].(map[string]interface{}); ok {
			if id, ok := data["id"].(float64); ok {
				createdDiaryID = uint(id)
				fmt.Printf("✅ 日记创建成功，ID: %d\n", createdDiaryID)
			}
		}
	}
}

func listDiaries() {
	fmt.Println("[6/11] 测试获取日记列表...")
	request("GET", "/api/diaries?page=1&page_size=10", &authToken, nil)
}

func searchDiary(keyword string) {
	fmt.Printf("[7/11] 测试搜索日记 (Keyword: %s)...\n", keyword)
	request("GET", "/api/diaries/search?q="+keyword+"&page=1&page_size=10", &authToken, nil)
}

func getDashboardStats() {
	fmt.Println("[8/11] 测试获取统计看板...")
	request("GET", "/api/stats/dashboard", &authToken, nil)
}

func getDiary(id uint) {
	fmt.Printf("[9/11] 测试获取日记详情 (ID: %d)...\n", id)
	request("GET", fmt.Sprintf("/api/diaries/%d", id), &authToken, nil)
}

func updateDiary(id uint, title, content string) {
	fmt.Printf("[10/11] 测试更新日记 (ID: %d)...\n", id)
	body := map[string]interface{}{
		"title":   title,
		"content": content,
		"weather": "Rainy", // 修改天气
	}
	request("PUT", fmt.Sprintf("/api/diaries/%d", id), &authToken, body)
}

func deleteDiary(id uint) {
	fmt.Printf("[11/11] 测试删除日记 (ID: %d)...\n", id)
	request("DELETE", fmt.Sprintf("/api/diaries/%d", id), &authToken, nil)
}

func request(method, path string, token *string, body interface{}) map[string]interface{} {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return nil
	}

	req.Header.Set("Content-Type", "application/json")
	if token != nil && *token != "" {
		req.Header.Set("Authorization", "Bearer "+*token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v (请确保服务器已启动)\n", err)
		return nil
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)

	fmt.Printf("➡️ %s %s Status: %d\n", method, path, resp.StatusCode)

	var result map[string]interface{}
	if err := json.Unmarshal(respBytes, &result); err == nil {
		// 简单打印一下 code 和 message
		if code, ok := result["code"].(float64); ok {
			msg, _ := result["message"].(string)
			fmt.Printf("   Code: %d, Msg: %s\n", int(code), msg)
		} else {
			fmt.Printf("   Resp: %s\n", string(respBytes))
		}
		return result
	} else {
		fmt.Printf("   Resp: %s\n", string(respBytes))
	}
	return nil
}

func printResponse(resp map[string]interface{}) {
	// 辅助打印，request里已经打印了
}
