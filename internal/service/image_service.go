package service

import (
	"bytes"
	"encoding/binary"
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

	face := basicfont.Face7x13
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{51, 51, 51, 255}),
		Face: face,
		Dot:  fixed.P(15, 25),
	}
	d.DrawString(title)

	d.Src = image.NewUniform(color.RGBA{153, 153, 153, 255})
	d.Dot = fixed.P(15, 45)
	d.DrawString(dateStr)

	d.Src = image.NewUniform(color.RGBA{51, 51, 51, 255})
	yPos := 75
	for _, line := range contentLines {
		if yPos > imgHeight-40 {
			break
		}
		d.Dot = fixed.P(15, yPos)
		d.DrawString(line)
		yPos += 20
	}

	if tagsStr != "" {
		d.Src = image.NewUniform(color.RGBA{24, 144, 255, 255})
		d.Dot = fixed.P(15, imgHeight-15)
		d.DrawString(tagsStr)
	}

	return encodeBMP(img)
}

func (s *ImageService) GeneratePlaceholderImage(message string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	setRect(img, 0, 0, imgWidth, imgHeight, color.RGBA{250, 250, 250, 255})

	lines := wrapText(message, 20)
	face := basicfont.Face7x13

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{153, 153, 153, 255}),
		Face: face,
	}

	yPos := imgHeight / 2
	startOffset := len(lines) * 10
	for i, line := range lines {
		d.Dot = fixed.P(20, yPos-startOffset+i*20)
		d.DrawString(line)
	}

	return encodeBMP(img)
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

func encodeBMP(img *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer

	fileHeader := make([]byte, 14)
	binary.LittleEndian.PutUint16(fileHeader[0:2], 0x4D42)

	rowSize := imgWidth * 4
	imageSize := uint32(rowSize * imgHeight)
	fileSize := uint32(14 + 40 + imageSize)

	binary.LittleEndian.PutUint32(fileHeader[2:6], fileSize)
	binary.LittleEndian.PutUint32(fileHeader[10:14], 14+40)

	infoHeader := make([]byte, 40)
	binary.LittleEndian.PutUint32(infoHeader[0:4], 40)
	binary.LittleEndian.PutInt32(infoHeader[4:8], int32(imgWidth))
	binary.LittleEndian.PutInt32(infoHeader[8:12], int32(imgHeight))
	binary.LittleEndian.PutUint16(infoHeader[12:14], 1)
	binary.LittleEndian.PutUint16(infoHeader[14:16], 32)
	binary.LittleEndian.PutUint32(infoHeader[20:24], 0)
	binary.LittleEndian.PutUint32(infoHeader[24:28], imageSize)

	buf.Write(fileHeader)
	buf.Write(infoHeader)

	pixels := make([]byte, imgWidth*imgHeight*4)
	for y := imgHeight - 1; y >= 0; y-- {
		for x := 0; x < imgWidth; x++ {
			c := img.RGBAAt(x, y)
			offset := (y*imgWidth + x) * 4
			pixels[offset] = c.B
			pixels[offset+1] = c.G
			pixels[offset+2] = c.R
			pixels[offset+3] = c.A
		}
	}
	buf.Write(pixels)

	return buf.Bytes(), nil
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
