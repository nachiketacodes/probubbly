package ratio

import "probubbly/internal/models"

const HouseCutPercent = 3.0

func Compute(yesCoins, noCoins int) models.RatioResult {
	total := yesCoins + noCoins

	// If nobody has predicted yet, both sides offer equal returns
	if total == 0 {
		return models.RatioResult{
			Yes:      1.94,
			No:       1.94,
			YesPct:   50,
			NoPct:    50,
			HouseCut: HouseCutPercent,
		}
	}

	yesFrac := float64(yesCoins) / float64(total)
	noFrac := float64(noCoins) / float64(total)
	base := 1.05

	// Crowded side gets lower return, underdog gets higher return
	yesRatio := (base + noFrac*3.5) / yesFrac
	noRatio := (base + yesFrac*3.5) / noFrac

	// Apply 3% house cut to both sides
	yesRatio = yesRatio * (1 - HouseCutPercent/100)
	noRatio = noRatio * (1 - HouseCutPercent/100)

	// Clamp between 1.02x and 9.6x
	yesRatio = clamp(yesRatio, 1.02, 9.6)
	noRatio = clamp(noRatio, 1.02, 9.6)

	return models.RatioResult{
		Yes:      round(yesRatio, 2),
		No:       round(noRatio, 2),
		YesPct:   int(yesFrac * 100),
		NoPct:    int(noFrac * 100),
		HouseCut: HouseCutPercent,
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func round(v float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int(v*pow+0.5)) / pow
}