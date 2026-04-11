package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
)

// FeedService 广场数据源服务
type FeedService struct{}

// NewFeedService 创建广场数据源服务
func NewFeedService() *FeedService {
	return &FeedService{}
}

// GetFeedSources 获取所有数据源
func (s *FeedService) GetFeedSources() ([]model.FeedSource, error) {
	db := store.GetDB()
	var sources []model.FeedSource
	if err := db.Order("created_at DESC").Find(&sources).Error; err != nil {
		return nil, err
	}
	return sources, nil
}

// CreateFeedSource 创建数据源
func (s *FeedService) CreateFeedSource(name, url string) (*model.FeedSource, error) {
	db := store.GetDB()
	source := model.FeedSource{
		Name:    name,
		URL:     url,
		Enabled: true,
	}
	if err := db.Create(&source).Error; err != nil {
		return nil, err
	}
	return &source, nil
}

// DeleteFeedSource 删除数据源
func (s *FeedService) DeleteFeedSource(id uint) error {
	db := store.GetDB()
	return db.Delete(&model.FeedSource{}, id).Error
}

// RemoteBibi 远程笔记（不存储，直接返回）
type RemoteBibi struct {
	ID            string      `json:"id"`
	Content       string      `json:"content"`
	LikeCount     int         `json:"like_count"`
	CommentCount  int         `json:"comment_count"`
	CreatedAt     string      `json:"created_at"`
	Creator       struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	} `json:"creator"`
	Tags          []RemoteTag   `json:"tags"`
	Comments      []RemoteComment `json:"comments"`
	SourceURL     string      `json:"source_url"`
}

// RemoteTag 远程标签
type RemoteTag struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// RemoteComment 远程评论
type RemoteComment struct {
	ID        uint   `json:"id"`
	ParentID  uint   `json:"parent_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Website   string `json:"website"`
	Content   string `json:"content"`
	Avatar    string `json:"avatar"`
	CreatedAt string `json:"created_at"`
}

// FetchBibisFromSource 从远程源获取笔记（不存储，直接返回）
func (s *FeedService) FetchBibisFromSource(sourceURL string) ([]RemoteBibi, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	baseURL := strings.TrimRight(sourceURL, "/")
	resp, err := client.Get(baseURL + "/api/v1/bibis?visibility=PUBLIC&page=1&page_size=50")
	if err != nil {
		return nil, fmt.Errorf("请求远程源失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("远程源返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result struct {
		Bibis []struct {
			ID           string `json:"id"`
			Content      string `json:"content"`
			LikeCount    int    `json:"like_count"`
			CommentCount int    `json:"comment_count"`
			CreatedAt    string `json:"created_at"`
			Creator      struct {
				Username string `json:"username"`
				Nickname string `json:"nickname"`
				Avatar   string `json:"avatar"`
			} `json:"creator"`
			Tags []struct {
				ID   uint   `json:"id"`
				Name string `json:"name"`
			} `json:"tags"`
			Comments []struct {
				ID        uint   `json:"id"`
				ParentID  uint   `json:"parent_id"`
				Name      string `json:"name"`
				Email     string `json:"email"`
				Website   string `json:"website"`
				Content   string `json:"content"`
				CreatedAt string `json:"created_at"`
			} `json:"comments"`
		} `json:"bibis"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	fmt.Printf("[DEBUG] Number of bibis: %d, first bibi comments: %d\n", len(result.Bibis), len(result.Bibis))

	bibis := make([]RemoteBibi, 0, len(result.Bibis))
	for _, b := range result.Bibis {
		rb := RemoteBibi{
			ID:           b.ID,
			Content:      b.Content,
			LikeCount:    b.LikeCount,
			CommentCount: b.CommentCount,
			CreatedAt:    b.CreatedAt,
			SourceURL:    sourceURL,
		}
		rb.Creator.Username = b.Creator.Username
		rb.Creator.Nickname = b.Creator.Nickname
		rb.Creator.Avatar = b.Creator.Avatar

		rb.Tags = make([]RemoteTag, len(b.Tags))
		for i, t := range b.Tags {
			rb.Tags[i] = RemoteTag{
				ID:   t.ID,
				Name: t.Name,
			}
		}

		rb.Comments = make([]RemoteComment, len(b.Comments))
		for i, c := range b.Comments {
			rb.Comments[i] = RemoteComment{
				ID:        c.ID,
				ParentID:  c.ParentID,
				Name:      c.Name,
				Email:     c.Email,
				Website:   c.Website,
				Content:   c.Content,
				CreatedAt: c.CreatedAt,
				Avatar:    model.GetGravatarURLWithSource(c.Email, GetGravatarSource()),
			}
		}

		bibis = append(bibis, rb)
	}

	return bibis, nil
}

// GetAllRemoteBibis 获取所有远程笔记（并行获取）
func (s *FeedService) GetAllRemoteBibis() ([]RemoteBibi, error) {
	db := store.GetDB()
	var sources []model.FeedSource
	if err := db.Where("enabled = ?", true).Find(&sources).Error; err != nil {
		return nil, err
	}

	if len(sources) == 0 {
		return []RemoteBibi{}, nil
	}

	var mu sync.Mutex
	allBibis := make([]RemoteBibi, 0)
	var wg sync.WaitGroup
	var fetchErr error

	for _, source := range sources {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			bibis, err := s.FetchBibisFromSource(url)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				fetchErr = err
				return
			}
			allBibis = append(allBibis, bibis...)
		}(source.URL)
	}

	wg.Wait()

	if len(allBibis) == 0 && fetchErr != nil {
		return nil, fetchErr
	}

	return allBibis, nil
}
