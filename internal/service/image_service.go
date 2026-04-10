package service

import (
	"fmt"
	"strings"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
)

const (
	imgWidth  = 400
	imgHeight = 300
)

type ImageService struct{}

func NewImageService() *ImageService {
	return &ImageService{}
}

func (s *ImageService) GetLatestPublicBibi() (*model.Bibi, error) {
	db := store.GetDB()
	var bibi model.Bibi
	err := db.Preload("Creator").Preload("Tags").
		Where("visibility = ?", "PUBLIC").
		Order("created_at DESC").
		First(&bibi).Error
	if err != nil {
		return nil, err
	}
	return &bibi, nil
}

func (s *ImageService) GenerateBibiCardImage(bibi *model.Bibi) ([]byte, error) {
	title := bibi.Creator.Nickname
	if title == "" {
		title = bibi.Creator.Username
	}
	dateStr := bibi.CreatedAt.Format("2006-01-02 15:04")
	content := stripMarkdown(bibi.Content)
	contentLines := wrapText(content, 28)

	var tagsStr string
	if len(bibi.Tags) > 0 {
		var tagNames []string
		for _, tag := range bibi.Tags {
			tagNames = append(tagNames, "#"+tag.Name)
		}
		tagsStr = strings.Join(tagNames, " ")
	}

	var contentSVG string
	for i, line := range contentLines {
		if i > 10 {
			break
		}
		contentSVG += fmt.Sprintf(`  <text x="15" y="%d" font-family="system-ui, -apple-system, 'PingFang SC', 'Microsoft YaHei', sans-serif" font-size="13" fill="#333">%s</text>`+"\n", 85+i*20, escapeXML(line))
	}

	var tagsSVG string
	if tagsStr != "" {
		tagsSVG = fmt.Sprintf(`  <text x="15" y="285" font-family="system-ui, -apple-system, 'PingFang SC', 'Microsoft YaHei', sans-serif" font-size="13" fill="#1890ff">%s</text>`, escapeXML(tagsStr))
	}

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="white"/>
  <rect x="0" y="0" width="%d" height="52" fill="#f5f5f5"/>
  <text x="15" y="25" font-family="system-ui, -apple-system, 'PingFang SC', 'Microsoft YaHei', sans-serif" font-size="14" font-weight="bold" fill="#333">%s</text>
  <text x="15" y="45" font-family="system-ui, -apple-system, 'PingFang SC', 'Microsoft YaHei', sans-serif" font-size="12" fill="#999">%s</text>
%s
%s
</svg>`, imgWidth, imgHeight, imgWidth, escapeXML(title), escapeXML(dateStr), contentSVG, tagsSVG)

	return []byte(svg), nil
}

func (s *ImageService) GeneratePlaceholderImage(message string) ([]byte, error) {
	lines := wrapText(message, 20)
	var contentSVG string
	startY := imgHeight / 2
	for i, line := range lines {
		contentSVG += fmt.Sprintf(`  <text x="200" y="%d" font-family="system-ui, -apple-system, 'PingFang SC', 'Microsoft YaHei', sans-serif" font-size="13" fill="#999" text-anchor="middle">%s</text>`+"\n", startY+i*22-len(lines)*11, escapeXML(line))
	}

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#fafafa"/>
%s
</svg>`, imgWidth, imgHeight, contentSVG)

	return []byte(svg), nil
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func stripMarkdown(text string) string {
	text = strings.ReplaceAll(text, "```\n", "")
	text = strings.ReplaceAll(text, "```", "")
	text = strings.ReplaceAll(text, "`", "")
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "_", "")
	text = strings.ReplaceAll(text, "# ", "")
	text = strings.ReplaceAll(text, "## ", "")
	text = strings.ReplaceAll(text, "### ", "")
	text = strings.ReplaceAll(text, "- ", "")
	text = strings.ReplaceAll(text, "* ", "")
	text = strings.ReplaceAll(text, "> ", "")
	text = strings.ReplaceAll(text, "\n\n", "\n")
	return strings.TrimSpace(text)
}

func wrapText(text string, maxChars int) []string {
	var result []string
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			result = append(result, "")
			continue
		}

		if len([]rune(line)) <= maxChars {
			result = append(result, line)
			continue
		}

		runes := []rune(line)
		var current string
		var currentLen int
		for _, r := range runes {
			if currentLen+1 > maxChars {
				result = append(result, current)
				current = string(r)
				currentLen = 1
			} else {
				current += string(r)
				currentLen++
			}
		}
		if current != "" {
			result = append(result, current)
		}
	}

	return result
}
