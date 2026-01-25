package builder

import (
	"math"
	"testing"
)

func TestMapFilter(t *testing.T) {
	in := []int{1, 2, 3, 4}
	out := Map(in, func(v int) int { return v * 2 })
	if len(out) != 4 || out[0] != 2 || out[3] != 8 {
		t.Fatalf("unexpected map output: %v", out)
	}

	filtered := Filter(out, func(v int) bool { return v%4 == 0 })
	if len(filtered) != 2 || filtered[0] != 4 || filtered[1] != 8 {
		t.Fatalf("unexpected filter output: %v", filtered)
	}
}

func TestTransformerSequence(t *testing.T) {
	t1 := func(v int) (int, error) { return v + 1, nil }
	t2 := func(v int) (int, error) { return v * 2, nil }
	seq := NewTransformerSequence(t1, t2)
	if len(seq) != 2 {
		t.Fatalf("expected 2 transformers, got %d", len(seq))
	}
}

func TestHaarWaveletRoundTrip(t *testing.T) {
	data := []float64{1, 1, 1, 1}
	hw := &HaarWavelet{}
	coeffs := hw.Transform(data)
	if len(coeffs) != len(data) {
		t.Fatalf("expected coeffs length %d, got %d", len(data), len(coeffs))
	}
	if math.Abs(coeffs[0]-math.Sqrt2) > 1e-9 {
		t.Fatalf("unexpected coeffs[0]: %.6f", coeffs[0])
	}
	out := hw.InverseTransform(coeffs)
	for i := range data {
		if math.Abs(out[i]-data[i]) > 1e-9 {
			t.Fatalf("round-trip mismatch at %d: got %.6f want %.6f", i, out[i], data[i])
		}
	}
}

func TestCompressFloatsRoundTrip(t *testing.T) {
	in := []float64{1.5, -2.0, 3.25}
	compressed, err := CompressFloats(in)
	if err != nil {
		t.Fatalf("CompressFloats error: %v", err)
	}
	if len(compressed) == 0 {
		t.Fatalf("expected compressed data")
	}

	out, err := DecompressFloats(compressed)
	if err != nil {
		t.Fatalf("DecompressFloats error: %v", err)
	}
	if len(out) != len(in) {
		t.Fatalf("expected %d items, got %d", len(in), len(out))
	}
	for i := range in {
		if math.Abs(out[i]-in[i]) > 1e-9 {
			t.Fatalf("decompressed mismatch at %d: got %.6f want %.6f", i, out[i], in[i])
		}
	}
}

func TestProcessWaveDataWithWavelet(t *testing.T) {
	wave := []complex128{
		complex(1, 0),
		complex(2, 0),
		complex(3, 0),
		complex(4, 0),
	}

	compressedHex, original, waveReal, decompressed, encoded, err := ProcessWaveDataWithWavelet(wave)
	if err != nil {
		t.Fatalf("ProcessWaveDataWithWavelet error: %v", err)
	}
	if compressedHex == "" || encoded == "" {
		t.Fatalf("expected encoded outputs to be non-empty")
	}
	if len(original) != len(wave) || len(waveReal) != len(wave) || len(decompressed) != len(wave) {
		t.Fatalf("unexpected output lengths")
	}
	for i := range wave {
		if original[i] != wave[i] {
			t.Fatalf("original wave mismatch at %d", i)
		}
		if waveReal[i] != real(wave[i]) {
			t.Fatalf("waveReal mismatch at %d", i)
		}
	}
}

func TestAnalyzeWaveOutputs(t *testing.T) {
	wave := []complex128{
		complex(1, 0),
		complex(0, 0),
		complex(-1, 0),
		complex(0, 0),
	}
	analysis := AnalyzeWave(wave, 100)

	ps, ok := analysis["power_spectrum"].([]float64)
	if !ok || len(ps) != len(wave)/2 {
		t.Fatalf("expected power_spectrum length %d, got %v", len(wave)/2, analysis["power_spectrum"])
	}

	dom, ok := analysis["dominant_frequency"].(float64)
	if !ok {
		t.Fatalf("expected dominant_frequency in analysis")
	}
	if dom < 0 || dom > 50 {
		t.Fatalf("unexpected dominant_frequency: %.2f", dom)
	}

	if _, ok := analysis["total_energy"].(float64); !ok {
		t.Fatalf("expected total_energy in analysis")
	}
	snr, ok := analysis["snr"].(float64)
	if !ok || math.IsNaN(snr) {
		t.Fatalf("expected valid snr in analysis")
	}
}
