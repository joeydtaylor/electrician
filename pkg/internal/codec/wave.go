package codec

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type WaveEncoder[T any] struct{}
type WaveDecoder[T any] struct{}

func (e *WaveEncoder[T]) Encode(w io.Writer, item T) error {
	wave, ok := any(item).(types.WaveData)
	if !ok {
		return fmt.Errorf("item is not of type WaveData")
	}

	// Encode fixed-size fields
	if err := binary.Write(w, binary.LittleEndian, int32(wave.ID)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, wave.DominantFreq); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, wave.TotalEnergy); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, wave.CompressionRatio); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, wave.MSE); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, wave.SNR); err != nil {
		return err
	}

	// Encode OriginalWave
	if err := binary.Write(w, binary.LittleEndian, int32(len(wave.OriginalWave))); err != nil {
		return err
	}
	for _, c := range wave.OriginalWave {
		if err := binary.Write(w, binary.LittleEndian, real(c)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, imag(c)); err != nil {
			return err
		}
	}

	// Encode CompressedHex
	if err := binary.Write(w, binary.LittleEndian, int32(len(wave.CompressedHex))); err != nil {
		return err
	}
	if _, err := w.Write([]byte(wave.CompressedHex)); err != nil {
		return err
	}

	// Encode FrequencyPeaks
	if err := binary.Write(w, binary.LittleEndian, int32(len(wave.FrequencyPeaks))); err != nil {
		return err
	}
	for _, peak := range wave.FrequencyPeaks {
		if err := binary.Write(w, binary.LittleEndian, peak.Freq); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, peak.Value); err != nil {
			return err
		}
	}

	return nil
}

func (d *WaveDecoder[T]) Decode(r io.Reader) (T, error) {
	var wave types.WaveData
	var item T

	// Decode fixed-size fields
	var id int32
	if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
		return item, err
	}
	wave.ID = int(id)

	if err := binary.Read(r, binary.LittleEndian, &wave.DominantFreq); err != nil {
		return item, err
	}
	if err := binary.Read(r, binary.LittleEndian, &wave.TotalEnergy); err != nil {
		return item, err
	}
	if err := binary.Read(r, binary.LittleEndian, &wave.CompressionRatio); err != nil {
		return item, err
	}
	if err := binary.Read(r, binary.LittleEndian, &wave.MSE); err != nil {
		return item, err
	}
	if err := binary.Read(r, binary.LittleEndian, &wave.SNR); err != nil {
		return item, err
	}

	// Decode OriginalWave
	var waveLen int32
	if err := binary.Read(r, binary.LittleEndian, &waveLen); err != nil {
		return item, err
	}
	wave.OriginalWave = make([]complex128, waveLen)
	for i := 0; i < int(waveLen); i++ {
		var real, imag float64
		if err := binary.Read(r, binary.LittleEndian, &real); err != nil {
			return item, err
		}
		if err := binary.Read(r, binary.LittleEndian, &imag); err != nil {
			return item, err
		}
		wave.OriginalWave[i] = complex(real, imag)
	}

	// Decode CompressedHex
	var hexLen int32
	if err := binary.Read(r, binary.LittleEndian, &hexLen); err != nil {
		return item, err
	}
	hexBytes := make([]byte, hexLen)
	if _, err := io.ReadFull(r, hexBytes); err != nil {
		return item, err
	}
	wave.CompressedHex = string(hexBytes)

	// Decode FrequencyPeaks
	var peaksLen int32
	if err := binary.Read(r, binary.LittleEndian, &peaksLen); err != nil {
		return item, err
	}
	wave.FrequencyPeaks = make([]types.Peak, peaksLen)
	for i := 0; i < int(peaksLen); i++ {
		if err := binary.Read(r, binary.LittleEndian, &wave.FrequencyPeaks[i].Freq); err != nil {
			return item, err
		}
		if err := binary.Read(r, binary.LittleEndian, &wave.FrequencyPeaks[i].Value); err != nil {
			return item, err
		}
	}

	return any(wave).(T), nil
}
