package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient 封装的HTTP客户端，支持连接池和统一的请求处理
type HTTPClient struct {
	client         *http.Client
	defaultTimeout time.Duration
}

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	MaxIdleConns        int           // 最大空闲连接数
	MaxIdleConnsPerHost int           // 每个主机的最大空闲连接数
	IdleConnTimeout     time.Duration // 空闲连接超时时间
	Timeout             time.Duration // 默认请求超时时间
	DisableKeepAlives   bool          // 是否禁用Keep-Alive
	ForceAttemptHTTP2   bool          // 是否强制尝试HTTP/2
}

// DefaultHTTPClientConfig 返回默认的HTTP客户端配置
func DefaultHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		Timeout:             30 * time.Second,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,
	}
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient(config *HTTPClientConfig) *HTTPClient {
	if config == nil {
		config = DefaultHTTPClientConfig()
	}

	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableKeepAlives:   config.DisableKeepAlives,
		ForceAttemptHTTP2:   config.ForceAttemptHTTP2,
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
		defaultTimeout: config.Timeout,
	}
}

// RequestOptions 请求选项
type RequestOptions struct {
	Method      string            // HTTP方法
	URL         string            // 请求URL
	Headers     map[string]string // 请求头
	Body        interface{}       // 请求体（会被序列化为JSON）
	Timeout     time.Duration     // 超时时间（可选，使用默认值如果为0）
	Context     context.Context   // 上下文（可选）
	BearerToken string            // Bearer Token（可选）
}

// Response 响应结构
type Response struct {
	StatusCode int         // 状态码
	Body       []byte      // 响应体
	Headers    http.Header // 响应头
}

// DoRequest 执行HTTP请求
func (c *HTTPClient) DoRequest(opts *RequestOptions) (*Response, error) {
	// 准备请求体
	var bodyReader io.Reader
	if opts.Body != nil {
		jsonData, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %v", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	// 创建上下文
	ctx := opts.Context
	if ctx == nil {
		timeout := opts.Timeout
		if timeout == 0 {
			timeout = c.defaultTimeout
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, opts.Method, opts.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	if opts.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+opts.BearerToken)
	}
	if opts.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	// 执行请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}

// DoJSONRequest 执行JSON请求并解析响应
func (c *HTTPClient) DoJSONRequest(opts *RequestOptions, result interface{}) error {
	resp, err := c.DoRequest(opts)
	if err != nil {
		return err
	}

	// 解析JSON响应
	if result != nil && len(resp.Body) > 0 {
		if err := json.Unmarshal(resp.Body, result); err != nil {
			return fmt.Errorf("解析响应失败: %v", err)
		}
	}

	return nil
}

// DoRequestWithStatus 执行请求并验证状态码
func (c *HTTPClient) DoRequestWithStatus(opts *RequestOptions, expectedStatus int) ([]byte, error) {
	resp, err := c.DoRequest(opts)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedStatus {
		return nil, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(resp.Body))
	}

	return resp.Body, nil
}

// DoRequestWithMultiStatus 执行请求并验证多个可能的状态码
func (c *HTTPClient) DoRequestWithMultiStatus(opts *RequestOptions, expectedStatuses ...int) ([]byte, int, error) {
	resp, err := c.DoRequest(opts)
	if err != nil {
		return nil, 0, err
	}

	// 检查状态码是否在预期列表中
	for _, status := range expectedStatuses {
		if resp.StatusCode == status {
			return resp.Body, resp.StatusCode, nil
		}
	}

	return nil, resp.StatusCode, fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(resp.Body))
}

// DoJSONRequestWithStatus 执行JSON请求、验证状态码并解析响应
func (c *HTTPClient) DoJSONRequestWithStatus(opts *RequestOptions, expectedStatus int, result interface{}) error {
	body, err := c.DoRequestWithStatus(opts, expectedStatus)
	if err != nil {
		return err
	}

	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("解析响应失败: %v", err)
		}
	}

	return nil
}

// Get 执行GET请求
func (c *HTTPClient) Get(url, bearerToken string) (*Response, error) {
	return c.DoRequest(&RequestOptions{
		Method:      http.MethodGet,
		URL:         url,
		BearerToken: bearerToken,
	})
}

// Post 执行POST请求
func (c *HTTPClient) Post(url, bearerToken string, body interface{}) (*Response, error) {
	return c.DoRequest(&RequestOptions{
		Method:      http.MethodPost,
		URL:         url,
		BearerToken: bearerToken,
		Body:        body,
	})
}

// Put 执行PUT请求
func (c *HTTPClient) Put(url, bearerToken string, body interface{}) (*Response, error) {
	return c.DoRequest(&RequestOptions{
		Method:      http.MethodPut,
		URL:         url,
		BearerToken: bearerToken,
		Body:        body,
	})
}

// Delete 执行DELETE请求
func (c *HTTPClient) Delete(url, bearerToken string) (*Response, error) {
	return c.DoRequest(&RequestOptions{
		Method:      http.MethodDelete,
		URL:         url,
		BearerToken: bearerToken,
	})
}

// GetJSON 执行GET请求并解析JSON响应
func (c *HTTPClient) GetJSON(url, bearerToken string, result interface{}) error {
	return c.DoJSONRequest(&RequestOptions{
		Method:      http.MethodGet,
		URL:         url,
		BearerToken: bearerToken,
	}, result)
}

// PostJSON 执行POST请求并解析JSON响应
func (c *HTTPClient) PostJSON(url, bearerToken string, body, result interface{}) error {
	return c.DoJSONRequest(&RequestOptions{
		Method:      http.MethodPost,
		URL:         url,
		BearerToken: bearerToken,
		Body:        body,
	}, result)
}

// PutJSON 执行PUT请求并解析JSON响应
func (c *HTTPClient) PutJSON(url, bearerToken string, body, result interface{}) error {
	return c.DoJSONRequest(&RequestOptions{
		Method:      http.MethodPut,
		URL:         url,
		BearerToken: bearerToken,
		Body:        body,
	}, result)
}

// GetClient 获取底层的http.Client（用于特殊情况）
func (c *HTTPClient) GetClient() *http.Client {
	return c.client
}
