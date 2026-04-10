package service

import (
	"bytes"
	"image"
	"image/color"
	"strings"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
	"golang.org/x/image/bmp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
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
	face := basicfont.Face7x13
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	setRect(img, 0, 0, imgWidth, imgHeight, color.RGBA{255, 255, 255, 255})
	setRect(img, 0, 0, imgWidth, 52, color.RGBA{245, 245, 245, 255})

	title := bibi.Creator.Nickname
	if title == "" {
		title = bibi.Creator.Username
	}
	dateStr := bibi.CreatedAt.Format("2006-01-02 15:04")

	content := stripMarkdown(bibi.Content)
	contentLines := wrapText(content, 26)

	var tagsStr string
	if len(bibi.Tags) > 0 {
		var tagNames []string
		for _, tag := range bibi.Tags {
			tagNames = append(tagNames, "#"+tag.Name)
		}
		tagsStr = strings.Join(tagNames, " ")
	}

	drawer := &font.Drawer{
		Dst: img,
		Face: face,
	}

	drawer.Src = image.NewUniform(color.RGBA{51, 51, 51, 255})
	drawer.Dot = fixed.P(15, 25)
	drawer.DrawString(title)

	drawer.Src = image.NewUniform(color.RGBA{153, 153, 153, 255})
	drawer.Dot = fixed.P(15, 45)
	drawer.DrawString(dateStr)

	drawer.Src = image.NewUniform(color.RGBA{51, 51, 51, 255})
	yPos := 75
	for _, line := range contentLines {
		if yPos > imgHeight-40 {
			break
		}
		drawer.Dot = fixed.P(15, yPos)
		drawer.DrawString(line)
		yPos += 20
	}

	if tagsStr != "" {
		drawer.Src = image.NewUniform(color.RGBA{24, 144, 255, 255})
		drawer.Dot = fixed.P(15, imgHeight-15)
		drawer.DrawString(tagsStr)
	}

	var buf bytes.Buffer
	if err := bmp.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *ImageService) GeneratePlaceholderImage(message string) ([]byte, error) {
	face := basicfont.Face7x13
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	setRect(img, 0, 0, imgWidth, imgHeight, color.RGBA{250, 250, 250, 255})

	lines := wrapText(message, 20)

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{153, 153, 153, 255}),
		Face: face,
	}

	yPos := imgHeight / 2
	startOffset := len(lines) * 10
	for i, line := range lines {
		drawer.Dot = fixed.P(20, yPos-startOffset+i*20)
		drawer.DrawString(line)
	}

	var buf bytes.Buffer
	if err := bmp.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func setRect(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if x >= 0 && x < imgWidth && y >= 0 && y < imgHeight {
				img.Set(x, y, c)
			}
		}
	}
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

		runes := []rune(line)
		if len(runes) <= maxChars {
			result = append(result, line)
			continue
		}

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
