package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"unicode/utf8"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
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
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	bgColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	clearRect(img, bgColor)

	headerBgColor := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	clearRectAt(img, 1, 1, imgWidth-1, 51, headerBgColor)

	face := basicfont.Face7x13
	black := color.RGBA{R: 51, G: 51, B: 51, A: 255}
	gray := color.RGBA{R: 153, G: 153, B: 153, A: 255}
	blue := color.RGBA{R: 24, G: 144, B: 255, A: 255}

	pointSize := fixed.I(13)
	fm := face.Metrics()
	baseline := pointSize.Ceil() - fm.Descent.Ceil()

	title := bibi.Creator.Nickname
	if title == "" {
		title = bibi.Creator.Username
	}
	drawStringAt(img, face, black, 15, baseline, title)

	dateStr := bibi.CreatedAt.Format("2006-01-02 15:04")
	drawStringAt(img, face, gray, 15, baseline+18, dateStr)

	content := stripMarkdown(bibi.Content)
	contentLines := wrapText(content, 48)
	yPos := baseline + 52
	for _, line := range contentLines {
		if yPos > imgHeight-35 {
			break
		}
		if len(line) > 0 {
			drawStringAt(img, face, black, 15, yPos, line)
		}
		yPos += 18
	}

	if len(bibi.Tags) > 0 {
		var tagNames []string
		for _, tag := range bibi.Tags {
			tagNames = append(tagNames, "#"+tag.Name)
		}
		tagsStr := strings.Join(tagNames, " ")
		drawStringAt(img, face, blue, 15, imgHeight-15, tagsStr)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *ImageService) GeneratePlaceholderImage(message string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	bgColor := color.RGBA{R: 250, G: 250, B: 250, A: 255}
	clearRect(img, bgColor)

	face := basicfont.Face7x13
	col := color.RGBA{R: 153, G: 153, B: 153, A: 255}

	pointSize := fixed.I(13)
	fm := face.Metrics()
	baseline := pointSize.Ceil() - fm.Descent.Ceil()

	lines := wrapText(message, 40)
	yPos := (imgHeight - len(lines)*18) / 2
	for _, line := range lines {
		drawStringAt(img, face, col, 20, yPos, line)
		yPos += 18
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func clearRect(img *image.RGBA, c color.Color) {
	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			img.Set(x, y, c)
		}
	}
}

func clearRectAt(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if x >= 0 && x < imgWidth && y >= 0 && y < imgHeight {
				img.Set(x, y, c)
			}
		}
	}
}

func drawStringAt(img *image.RGBA, face font.Face, c color.Color, x, y int, s string) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(s)
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

		words := strings.Fields(line)
		var current strings.Builder
		count := 0

		for _, word := range words {
			wordLen := utf8.RuneCountInString(word)
			if current.Len() == 0 {
				current.WriteString(word)
				count = wordLen
			} else if count+1+wordLen <= maxChars {
				current.WriteString(" ")
				current.WriteString(word)
				count += 1 + wordLen
			} else {
				result = append(result, current.String())
				current.Reset()
				current.WriteString(word)
				count = wordLen
			}
		}

		if current.Len() > 0 {
			result = append(result, current.String())
		}
	}

	return result
}
