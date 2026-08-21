package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

type BenchmarkResult struct {
	Model        string
	Prompt       string
	Duration     time.Duration
	StatusCode   int
	PayloadBytes int
	Success      bool
	ErrorMsg     string
}

func TestImageGenerationBenchmark(t *testing.T) {
	_ = godotenv.Load(".env")

	prompts := []string{
		"cute 2d cartoon apple with happy smiling face, isolated on pure solid white background, vector clipart for kids, clean lines",
		"a cute cartoon cat sitting, isolated on pure solid white background, flat colorful clipart for kindergarten",
		"three ripe yellow bananas, isolated on pure white background, simple cartoon illustration for elementary worksheet",
	}

	client := &http.Client{
		Timeout: 45 * time.Second,
	}

	results := []BenchmarkResult{}

	fmt.Println("\n========================================================")
	fmt.Println("🚀 RUNNING LIVE IMAGE GENERATION BENCHMARK (AJARVISUAL)")
	fmt.Println("========================================================")

	for idx, prompt := range prompts {
		fmt.Printf("\n--- Test Case %d: \"%s\" ---\n", idx+1, prompt[:35]+"...")

		// 1. Test Pollinations.ai (Default Turbo Engine)
		t0 := time.Now()
		pollTurboURL := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?width=512&height=512&nologo=true&seed=%d", url.QueryEscape(prompt), 12345+idx)
		
		reqTurbo, _ := http.NewRequestWithContext(context.Background(), "GET", pollTurboURL, nil)
		respTurbo, errTurbo := client.Do(reqTurbo)
		durTurbo := time.Since(t0)

		resTurbo := BenchmarkResult{
			Model:    "Pollinations (Default/Turbo)",
			Prompt:   prompt,
			Duration: durTurbo,
		}

		if errTurbo != nil {
			resTurbo.Success = false
			resTurbo.ErrorMsg = errTurbo.Error()
			fmt.Printf("❌ [Pollinations Turbo] Error: %v (%v)\n", errTurbo, durTurbo)
		} else {
			defer respTurbo.Body.Close()
			body, _ := io.ReadAll(respTurbo.Body)
			resTurbo.StatusCode = respTurbo.StatusCode
			resTurbo.PayloadBytes = len(body)
			resTurbo.Success = respTurbo.StatusCode == http.StatusOK
			fmt.Printf("✅ [Pollinations Turbo] Status: %d | Size: %.1f KB | Time: %v\n", respTurbo.StatusCode, float64(len(body))/1024.0, durTurbo)
		}
		results = append(results, resTurbo)

		// 2. Test Pollinations.ai (FLUX.1 Engine)
		t1 := time.Now()
		pollFluxURL := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?model=flux&width=512&height=512&nologo=true&seed=%d", url.QueryEscape(prompt), 12345+idx)
		
		reqFlux, _ := http.NewRequestWithContext(context.Background(), "GET", pollFluxURL, nil)
		respFlux, errFlux := client.Do(reqFlux)
		durFlux := time.Since(t1)

		resFlux := BenchmarkResult{
			Model:    "Pollinations (FLUX.1 Engine)",
			Prompt:   prompt,
			Duration: durFlux,
		}

		if errFlux != nil {
			resFlux.Success = false
			resFlux.ErrorMsg = errFlux.Error()
			fmt.Printf("❌ [Pollinations FLUX] Error: %v (%v)\n", errFlux, durFlux)
		} else {
			defer respFlux.Body.Close()
			body, _ := io.ReadAll(respFlux.Body)
			resFlux.StatusCode = respFlux.StatusCode
			resFlux.PayloadBytes = len(body)
			resFlux.Success = respFlux.StatusCode == http.StatusOK
			fmt.Printf("✅ [Pollinations FLUX] Status: %d | Size: %.1f KB | Time: %v\n", respFlux.StatusCode, float64(len(body))/1024.0, durFlux)
		}
		results = append(results, resFlux)
	}

	fmt.Println("\n========================================================")
	fmt.Println("📊 BENCHMARK SUMMARY TABLE")
	fmt.Println("========================================================")
	fmt.Printf("%-32s | %-12s | %-12s | %-12s\n", "Model", "Avg Latency", "Avg Size", "Success Rate")
	fmt.Println("---------------------------------------------------------------------------------")

	// Calculate averages
	var turboTimeTotal, fluxTimeTotal time.Duration
	var turboSizeTotal, fluxSizeTotal int
	var turboSuccessCount, fluxSuccessCount int
	var count = len(prompts)

	for _, r := range results {
		if r.Model == "Pollinations (Default/Turbo)" {
			turboTimeTotal += r.Duration
			turboSizeTotal += r.PayloadBytes
			if r.Success {
				turboSuccessCount++
			}
		} else {
			fluxTimeTotal += r.Duration
			fluxSizeTotal += r.PayloadBytes
			if r.Success {
				fluxSuccessCount++
			}
		}
	}

	fmt.Printf("%-32s | %-12v | %-10.1f KB | %-12.0f%%\n",
		"Pollinations (Default/Turbo)",
		turboTimeTotal/time.Duration(count),
		float64(turboSizeTotal)/(float64(count)*1024.0),
		(float64(turboSuccessCount)/float64(count))*100.0)

	fmt.Printf("%-32s | %-12v | %-10.1f KB | %-12.0f%%\n",
		"Pollinations (FLUX.1 Engine)",
		fluxTimeTotal/time.Duration(count),
		float64(fluxSizeTotal)/(float64(count)*1024.0),
		(float64(fluxSuccessCount)/float64(count))*100.0)
	fmt.Println("========================================================")
}
