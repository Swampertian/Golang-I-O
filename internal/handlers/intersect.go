package handlers

import (
	"context"
	"encoding/json"
	"fire-go/internal/db"
	"fire-go/internal/logger"
	"fire-go/internal/utils"
	"net/http"
	"strconv"
	"time"
	"unsafe"

	"log/slog"
)

func logMemoryUsage(vars map[string]interface{}) {
	for name, value := range vars {
		size := unsafe.Sizeof(value)
		logger.Log.Debug("memory_usage",
			slog.String("variable", name),
			slog.Uint64("bytes", uint64(size)),
		)
	}
}

type FireResponse struct {
	ID             int64           `json:"id"`
	Type           string          `json:"type"`
	MunicipalityID int64           `json:"municipality_id"`
	Year           int             `json:"year"`
	Month          int             `json:"month"`
	AreaHa         float64         `json:"area_ha"`
	Geom           json.RawMessage `json:"geom"`
}

func GetFireIntersectFiltered(w http.ResponseWriter, r *http.Request) {
	totalStart := time.Now()

	ctx := context.Background()

	// ----------------------------
	// Extração e parsing dos filtros
	// ----------------------------
	dataInicioStr := r.URL.Query().Get("dataInicio")
	dataFimStr := r.URL.Query().Get("dataFim")
	codigoStr := r.URL.Query().Get("codigo")

	var where []string
	var params []interface{}

	if dataInicioStr != "" {
		where = append(where, "date >= $"+strconv.Itoa(len(params)+1))
		t, _ := time.Parse("2006-01-02", dataInicioStr)
		params = append(params, t)
	}

	if dataFimStr != "" {
		where = append(where, "date <= $"+strconv.Itoa(len(params)+1))
		t, _ := time.Parse("2006-01-02", dataFimStr)
		params = append(params, t)
	}

	if codigoStr != "" {
		where = append(where, "municipality_id = $"+strconv.Itoa(len(params)+1))
		params = append(params, codigoStr)
	}

	// Log de memória dos parâmetros iniciais
	logMemoryUsage(map[string]interface{}{
		"dataInicioStr": dataInicioStr,
		"dataFimStr":    dataFimStr,
		"codigoStr":     codigoStr,
		"where_slice":   where,
		"params_slice":  params,
	})

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + joinWhere(where)
	}

	query := `
		SELECT
			id,
			type,
			municipality_id,
			year,
			month,
			area_ha,
			geom
		FROM fire_intersect
		` + whereClause

	// Log de memória após construção da query
	logMemoryUsage(map[string]interface{}{
		"whereClause": whereClause,
		"query":       query,
	})

	// ----------------------------
	// Tempo da consulta SQL
	// ----------------------------
	dbStart := time.Now()
	rows, err := db.Pool.Query(ctx, query, params...)
	dbDuration := time.Since(dbStart)

	if err != nil {
		logger.Log.Error("db_error",
			slog.String("error", err.Error()),
			slog.String("query", query),
		)
		http.Error(w, "Erro ao executar query: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	logger.Log.Info("db_query_execution",
		slog.String("query", query),
		slog.Int("params_count", len(params)),
		slog.Int64("duration_ms", dbDuration.Milliseconds()),
	)

	// ----------------------------
	// Leitura das linhas + tempo de conversão EWKB → GeoJSON
	// ----------------------------
	var list []FireResponse
	var conversionTime time.Duration

	for rows.Next() {
		var (
			id             int64
			typeField      string
			municipalityID int64
			year           int
			month          int
			areaHa         float64
			geomWKB        string
		)

		if err := rows.Scan(&id, &typeField, &municipalityID, &year, &month, &areaHa, &geomWKB); err != nil {
			logger.Log.Error("row_scan_error", "error", err)
			http.Error(w, err.Error(), 500)
			return
		}

		logMemoryUsage(map[string]interface{}{
			"id":             id,
			"typeField":      typeField,
			"municipalityID": municipalityID,
			"year":           year,
			"month":          month,
			"areaHa":         areaHa,
			"geomWKB":        geomWKB,
		})

		// Tempo de conversão geométrica
		convStart := time.Now()
		geojson, err := utils.EWKBHexToGeoJSON(geomWKB)
		convDuration := time.Since(convStart)
		conversionTime += convDuration

		if err != nil {
			logger.Log.Error("geom_conversion_error",
				slog.String("error", err.Error()),
			)
			http.Error(w, "Erro convertendo geom: "+err.Error(), 500)
			return
		}

		list = append(list, FireResponse{
			ID:             id,
			Type:           typeField,
			MunicipalityID: municipalityID,
			Year:           year,
			Month:          month,
			AreaHa:         areaHa,
			Geom:           json.RawMessage(geojson),
		})
	}

	logger.Log.Info("geom_conversion_summary",
		slog.Int("records", len(list)),
		slog.Int64("duration_ms", conversionTime.Milliseconds()),
	)

	serStart := time.Now()
	payload, err := json.Marshal(list)
	serDuration := time.Since(serStart)

	if err != nil {
		logger.Log.Error("serialization_error", "error", err)
		http.Error(w, err.Error(), 500)
		return
	}

	logMemoryUsage(map[string]interface{}{
		"list_slice": list,
		"payload":    payload,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)

	totalDuration := time.Since(totalStart)
	logger.Log.Info("request_complete",
		slog.String("endpoint", "/fire_intersect"),
		slog.Int("records", len(list)),
		slog.Int("response_bytes", len(payload)),
		slog.Int64("db_ms", dbDuration.Milliseconds()),
		slog.Int64("geom_conversion_ms", conversionTime.Milliseconds()),
		slog.Int64("serialization_ms", serDuration.Milliseconds()),
		slog.Int64("total_ms", totalDuration.Milliseconds()),
	)
}

func joinWhere(parts []string) string {
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += " AND " + parts[i]
	}
	return out
}
