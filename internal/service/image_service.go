package service

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/bmp"
)

const (
	imgWidth  = 400
	imgHeight = 300
	fontPath  = "/app/fonts/MiSans-Regular.ttf"
)

var loadedFont *truetype.Font
var ftCtx *freetype.Context

func init() {
	os.MkdirAll(filepath.Dir(fontPath), 0755)
}

func init() {
	os.MkdirAll(filepath.Dir(fontPath), 0755)
}

func loadFont() (*truetype.Font, error) {
	if loadedFont != nil {
		return loadedFont, nil
	}

	f, err := os.ReadFile(fontPath)
	if err != nil {
		fmt.Printf("loadFont: failed to read font file: %v\n", err)
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}

	ttfFont, err := truetype.Parse(f)
	if err != nil {
		fmt.Printf("loadFont: failed to parse font: %v\n", err)
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	loadedFont = ttfFont

	ftCtx = freetype.NewContext()
	ftCtx.SetFont(ttfFont)
	ftCtx.SetDPI(72)

	fmt.Println("loadFont: success")
	return loadedFont, nil
}

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

func drawText(img *image.RGBA, text string, x, y int, col color.Color, size float64, ftFont *truetype.Font) {
	if ftFont == nil || ftCtx == nil {
		return
	}
	ftCtx.SetDst(img)
	ftCtx.SetSrc(image.NewUniform(col))
	ftCtx.SetFont(ftFont)
	ftCtx.SetFontSize(size)
	ftCtx.SetClip(img.Bounds())
	_, err := ftCtx.DrawString(text, freetype.Pt(x, y))
	if err != nil {
		fmt.Printf("drawText error: %v\n", err)
	}
}

func (s *ImageService) GenerateBibiCardImage(bibi *model.Bibi) ([]byte, error) {
	ftFont, err := loadFont()
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	setRect(img, 0, 0, imgWidth, imgHeight, color.RGBA{255, 255, 255, 255})
	setRect(img, 0, 0, imgWidth, 52, color.RGBA{245, 245, 245, 255})

	title := bibi.Creator.Nickname
	if title == "" {
		title = bibi.Creator.Username
	}
	dateStr := bibi.CreatedAt.Format("2006-01-02 15:04")

	content := stripMarkdown(bibi.Content)
	contentLines := wrapText(content, 22)

	var tagsStr string
	if len(bibi.Tags) > 0 {
		var tagNames []string
		for _, tag := range bibi.Tags {
			tagNames = append(tagNames, "#"+tag.Name)
		}
		tagsStr = strings.Join(tagNames, " ")
	}

	drawText(img, title, 15, 25, color.RGBA{51, 51, 51, 255}, 14, ftFont)
	drawText(img, dateStr, 15, 45, color.RGBA{153, 153, 153, 255}, 12, ftFont)

	yPos := 75
	for _, line := range contentLines {
		if yPos > imgHeight-40 {
			break
		}
		drawText(img, line, 15, yPos, color.RGBA{51, 51, 51, 255}, 14, ftFont)
		yPos += 20
	}

	if tagsStr != "" {
		drawText(img, tagsStr, 15, imgHeight-15, color.RGBA{24, 144, 255, 255}, 12, ftFont)
	}

	var buf bytes.Buffer
	err = bmp.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *ImageService) GenerateRemoteBibiCardImage(rb *RemoteBibi) ([]byte, error) {
	ftFont, err := loadFont()
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	setRect(img, 0, 0, imgWidth, imgHeight, color.RGBA{255, 255, 255, 255})
	setRect(img, 0, 0, imgWidth, 52, color.RGBA{245, 245, 245, 255})

	title := rb.Creator.Nickname
	if title == "" {
		title = rb.Creator.Username
	}
	dateStr := rb.CreatedAt

	content := stripMarkdown(rb.Content)
	contentLines := wrapText(content, 22)

	var tagsStr string
	if len(rb.Tags) > 0 {
		var tagNames []string
		for _, tag := range rb.Tags {
			tagNames = append(tagNames, "#"+tag.Name)
		}
		tagsStr = strings.Join(tagNames, " ")
	}

	drawText(img, title, 15, 25, color.RGBA{51, 51, 51, 255}, 14, ftFont)
	drawText(img, dateStr, 15, 45, color.RGBA{153, 153, 153, 255}, 12, ftFont)

	yPos := 75
	for _, line := range contentLines {
		if yPos > imgHeight-40 {
			break
		}
		drawText(img, line, 15, yPos, color.RGBA{51, 51, 51, 255}, 14, ftFont)
		yPos += 20
	}

	if tagsStr != "" {
		drawText(img, tagsStr, 15, imgHeight-15, color.RGBA{24, 144, 255, 255}, 12, ftFont)
	}

	var buf bytes.Buffer
	err = bmp.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *ImageService) EncodeTestPattern(img *image.RGBA) []byte {
	var buf bytes.Buffer
	bmp.Encode(&buf, img)
	return buf.Bytes()
}

func (s *ImageService) GeneratePlaceholderImage(message string) ([]byte, error) {
	ftFont, err := loadFont()
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	setRect(img, 0, 0, imgWidth, imgHeight, color.RGBA{250, 250, 250, 255})

	lines := wrapText(message, 20)

	yPos := imgHeight / 2
	startOffset := len(lines) * 10
	for i, line := range lines {
		drawText(img, line, 20, yPos-startOffset+i*20, color.RGBA{153, 153, 153, 255}, 14, ftFont)
	}

	var buf bytes.Buffer
	err = bmp.Encode(&buf, img)
	if err != nil {
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

func encode1BitBMP(img *image.RGBA) []byte {
	width := imgWidth
	height := imgHeight
	rowBytes := 50

	pixelData := make([]byte, rowBytes*height)

	for y := height - 1; y >= 0; y-- {
		rowIdx := (height - 1 - y) * rowBytes
		for x := 0; x < width; x++ {
			byteIdx := rowIdx + x/8
			bitIdx := 7 - (x % 8)
			r, g, b, a := img.At(x, y).RGBA()
			luminance := (r + g + b) / 3
			if a > 32768 && luminance > 40000 {
				pixelData[byteIdx] |= (1 << bitIdx)
			}
		}
	}

	return pixelData
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
