package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/app/assets"
	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/app/request"
	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/app/response"
	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/color"
	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/files"
	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/roblox"
)

var CompatiblePluginVersion = ""

func getOutputFileName(reuploadType string) string {
	t := time.Now()
	return fmt.Sprintf("Output_%s_%s.json", reuploadType, t.Format("2006-01-02_15-04-05"))
}

func writeResponse(w http.ResponseWriter, resp *response.Response, busy bool, finished bool, isFinal bool) {
	if isFinal {
		if resp.Len() > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := resp.EncodeJSON(json.NewEncoder(w)); err != nil {
				log.Fatal(err)
			} else {
				resp.Clear()
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "done")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if resp.Len() == 0 && !busy {
		_, _ = fmt.Fprint(w, "[]")
		return
	}

	if err := resp.EncodeJSON(json.NewEncoder(w)); err != nil {
		log.Fatal(err)
	} else {
		resp.Clear()
	}
}

func serve(c *roblox.Client) error {
	var exportedJSONName string
	var exportJSON bool
	var busy bool
	finished := true

	respHistory := make([]response.ResponseItem, 0)
	resp := response.New(func(i response.ResponseItem) {
		if exportJSON {
			respHistory = append(respHistory, i)

			j, err := json.Marshal(respHistory)
			if err != nil {
				log.Fatal(err)
			}

			if err := files.Write(exportedJSONName, string(j)); err != nil {
				log.Fatal(err)
			}
		}
	})

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if resp.Len() == 0 && !busy {
			if !finished {
				finished = true
				busy = false
				exportJSON = false

				resp.Clear()
				respHistory = make([]response.ResponseItem, 0)

				fmt.Println("Finished reuploading. (you can rerun without restarting)")
				writeResponse(w, resp, busy, finished, true)
				return
			}

			writeResponse(w, resp, busy, finished, false)
			return
		}

		writeResponse(w, resp, busy, finished, false)
	})

	http.HandleFunc("POST /reupload", func(w http.ResponseWriter, r *http.Request) {
		if busy || !finished {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		var req request.RawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			color.Error.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if CompatiblePluginVersion != "" && req.PluginVersion != CompatiblePluginVersion {
			w.WriteHeader(http.StatusConflict)
			return
		}

		if exists := assets.DoesModuleExist(req.AssetType); !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		startReupload, err := assets.NewReuploadHandlerWithType(req.AssetType, c, &req, resp)
		if err != nil {
			color.Error.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if exportJSON = req.ExportJSON; exportJSON {
			exportedJSONName = getOutputFileName(req.AssetType)
		}

		busy = true
		finished = false

		go func() {
			start := time.Now()
			err := startReupload()
			busy = false
			if err != nil {
				finished = true
				color.Error.Println("Failed to start reuploading: ", err)
				return
			}

			duration := time.Since(start)
			fmt.Printf("Reuploading took %d hours, %d minutes, and %d seconds\n", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)
			fmt.Println("Waiting for client to finish changing ids...")
		}()

		w.WriteHeader(http.StatusOK)
	})

	return http.ListenAndServe(":"+port, nil)
}

