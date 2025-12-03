package handlers

import (
	"context"
	"encoding/json"
	"fire-go/internal/db"
	"fire-go/internal/utils"
	"net/http"
	"strconv"
	"time"
)

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
	ctx := context.Background()

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

	rows, err := db.Pool.Query(ctx, query, params...)
	if err != nil {
		http.Error(w, "Erro ao executar query: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	var list []FireResponse

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

		err := rows.Scan(
			&id,
			&typeField,
			&municipalityID,
			&year,
			&month,
			&areaHa,
			&geomWKB,
		)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		geojson, err := utils.EWKBHexToGeoJSON(geomWKB)
		if err != nil {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func joinWhere(parts []string) string {
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += " AND " + parts[i]
	}
	return out
}
