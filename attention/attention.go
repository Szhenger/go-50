package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strings"

	"gocv.io/x/gocv"
	"github.com/sugarme/gotorch" // Standard LibTorch Go binding architecture
)

// ── Hyperparameters & Constants ─────────────────────────────────────────────

const (
	K             = 3
	GridSize      = 40
	PixelsPerWord = 200
	MaskTokenID   = 103 // Default [MASK] id for bert-base-uncased
)

// ── Mock Tokenizer ──────────────────────────────────────────────────────────

type Tokenizer struct {
	MaskTokenID int64
	MaskToken   string
}

func NewTokenizer() Tokenizer {
	return Tokenizer{
		MaskTokenID: MaskTokenID,
		MaskToken:   "[MASK]",
	}
}

func (t Tokenizer) Encode(text string) []int64 {
	// Mock tokenization logic matching C++ stub
	return []int64{101, 2023, 2003, 103, 1012, 102} // [CLS] this is [MASK] . [SEP]
}

func (t Tokenizer) GetTokens(text string) []string {
	return []string{"[CLS]", "this", "is", "[MASK]", ".", "[SEP]"}
}

func (t Tokenizer) Decode(tokenID int64) string {
	return "example" // Mock decoded token
}

// ── Helper Functions ────────────────────────────────────────────────────────

func getMaskTokenIndex(maskTokenID int64, inputIDs []int64) (int, bool) {
	for i, id := range inputIDs {
		if id == maskTokenID {
			return i, true
		}
	}
	return 0, false
}

func getColorForAttentionScore(attentionScore float32) color.RGBA {
	colorScore := uint8(math.Round(255.0 * float64(attentionScore)))
	// Return grayscale representation (BGR/RGBA channels assigned identically)
	return color.RGBA{R: colorScore, G: colorScore, B: colorScore, A: 255}
}

// ── Attention Visualization Logic ───────────────────────────────────────────

func generateDiagram(layerNumber, headNumber int, tokens []string, attentionWeights gotorch.Tensor) {
	imageSize := GridSize*len(tokens) + PixelsPerWord

	// Create a new black matrix image (8-bit unsigned, 3 channels)
	img := gocv.NewMatWithSize(imageSize, imageSize, gocv.MatTypeCV8UC3)
	defer img.Close()
	img.SetTo(gocv.NewScalar(0, 0, 0, 0))

	fontFace := gocv.FontHersheySimplex
	fontScale := 0.8
	thickness := 1
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Draw each token onto the image canvas
	for i, token := range tokens {
		// Draw token columns (Rotated 90 degrees counter-clockwise)
		textSize := gocv.GetTextSize(token, fontFace, fontScale, thickness)

		// Create temporary canvas buffer for text rotation
		tempTextImg := gocv.NewMatWithSize(textSize.Y+10, textSize.X, gocv.MatTypeCV8UC3)
		tempTextImg.SetTo(gocv.NewScalar(0, 0, 0, 0))
		gocv.PutText(&tempTextImg, token, image.Pt(0, textSize.Y), fontFace, fontScale, white, thickness)

		textImg := gocv.NewMat()
		gocv.Rotate(tempTextImg, &textImg, gocv.Rotate90Counterclockwise)
		tempTextImg.Close()

		colX := imageSize - PixelsPerWord
		colY := PixelsPerWord + i*GridSize
		colRoi := image.Rect(colX, colY, colX+textImg.Cols(), colY+textImg.Rows())

		if colRoi.In(image.Rect(0, 0, img.Cols(), img.Rows())) {
			subMat := img.Region(colRoi)
			textImg.CopyTo(&subMat)
			subMat.Close()
		}
		textImg.Close()

		// Draw sequential token rows
		rowX := PixelsPerWord - textSize.X - 10 // Padding offset
		rowY := PixelsPerWord + i*GridSize + (GridSize+textSize.Y)/2
		gocv.PutText(&img, token, image.Pt(rowX, rowY), fontFace, fontScale, white, thickness)
	}

	// Draw each word's attention heatmap grid
	for i := 0; i < len(tokens); i++ {
		y := PixelsPerWord + i*GridSize
		for j := 0; j < len(tokens); j++ {
			x := PixelsPerWord + j*GridSize

			// Extract coordinate value safely from LibTorch Tensor
			weight := attentionWeights.At(i, j)
			cellColor := getColorForAttentionScore(weight)

			// Draw filled grid cell
			cell := image.Rect(x, y, x+GridSize, y+GridSize)
			gocv.Rectangle(&img, cell, cellColor, -1) // -1 signifies a filled element
		}
	}

	filename := fmt.Sprintf("Attention_Layer%d_Head%d.png", layerNumber+1, headNumber+1)
	gocv.IMWrite(filename, img)
}

func visualizeAttentions(tokens []string, attentions gotorch.Tensor) {
	// Shape expected: [Num_Layers, Batch_Size, Num_Heads, Seq_Len, Seq_Len]
	numLayers := attentions.Size(0)
	numHeads := attentions.Size(2)

	for i := 0; i < numLayers; i++ {
		for j := 0; j < numHeads; j++ {
			// Extract 2D Attention Weight Slice [i, Batch 0, j]
			headAttention := attentions.Select(0, i).Select(0, 0).Select(0, j).Cpu()
			generateDiagram(i, j, tokens, headAttention)
		}
	}
}

// ── Main Entrypoint ─────────────────────────────────────────────────────────

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Text: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	text = strings.TrimSpace(text)

	tokenizer := NewTokenizer()

	// Tokenize input string
	inputIDsVec := tokenizer.Encode(text)
	tokens := tokenizer.GetTokens(text)

	maskTokenIndex, found := getMaskTokenIndex(tokenizer.MaskTokenID, inputIDsVec)
	if !found {
		fmt.Fprintf(os.Stderr, "Input must include mask token %s.\n", tokenizer.MaskToken)
		os.Exit(1)
	}

	// Convert input array to LibTorch Int64 Tensor and insert Batch dimension
	inputTensor := gotorch.NewTensor(inputIDsVec).Unsqueeze(0)

	// Load pre-trained script model export
	model, err := gotorch.Load("bert_masked_lm.pt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading the model. Ensure 'bert_masked_lm.pt' exists in the directory.")
		os.Exit(1)
	}

	// Run forward execution pass
	outputTuple, err := model.Forward(inputTensor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Inference processing failure: %v\n", err)
		os.Exit(1)
	}

	logits := outputTuple.Get(0)     // Shape: [1, Seq_Len, Vocab_Size]
	attentions := outputTuple.Get(1) // Shape: [Layers, 1, Heads, Seq, Seq]

	// Isolate predictions for the specific [MASK] target token position
	maskTokenLogits := logits.Select(0, 0).Select(0, maskTokenIndex)

	// Retrieve Top-K target evaluation candidates
	topKValues, topKIndices := maskTokenLogits.TopK(K, -1, true, true)

	for i := 0; i < K; i++ {
		predictedID := topKIndices.At(i)
		decodedToken := tokenizer.Decode(predictedID)

		// Mirror string swap replacement logic
		predictedText := strings.Replace(text, tokenizer.MaskToken, decodedToken, 1)
		fmt.Println(predictedText)
	}

	// Process and export structural image attention matrices
	visualizeAttentions(tokens, attentions)
}
