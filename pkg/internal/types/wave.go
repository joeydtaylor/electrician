package types

type WaveData struct {
	ID               int
	OriginalWave     []complex128
	CompressedHex    string
	DominantFreq     float64
	TotalEnergy      float64
	CompressionRatio float64
	MSE              float64
	FrequencyPeaks   []Peak
	SNR              float64
}

type Peak struct {
	Freq  float64
	Value float64
}
