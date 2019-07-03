package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/xubiosueldos/concepto/structConcepto"

	"github.com/xubiosueldos/framework/configuracion"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/xubiosueldos/autenticacion/apiclientautenticacion"
	"github.com/xubiosueldos/conexionBD/apiclientconexionbd"
	"github.com/xubiosueldos/framework"
	"github.com/xubiosueldos/novedad/structNovedad"
	"github.com/xubiosueldos/novedad/structNovedadMin"
	_ "github.com/xubiosueldos/novedad/structNovedadMin"
)

var nombreMicroservicio string = "novedad"

// Sirve para controlar si el server esta OK
func Healthy(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Healthy"))
}

func NovedadList(w http.ResponseWriter, r *http.Request) {

	tokenValido, tokenAutenticacion := apiclientautenticacion.CheckTokenValido(w, r)
	if tokenValido {
		versionMicroservicio := obtenerVersionNovedad()
		tenant := apiclientautenticacion.ObtenerTenant(tokenAutenticacion)

		db := apiclientconexionbd.ObtenerDB(tenant, nombreMicroservicio, versionMicroservicio, AutomigrateTablasPrivadas)

		defer db.Close()

		var novedades []structNovedadMin.Novedad

		db.Set("gorm:auto_preload", true).Find(&novedades)

		for i, novedad := range novedades {
			concepto := obtenerConcepto(*novedad.Conceptoid, r)
			novedades[i].Concepto = concepto
		}

		framework.RespondJSON(w, http.StatusOK, novedades)
	}

}

func NovedadShow(w http.ResponseWriter, r *http.Request) {

	tokenValido, tokenAutenticacion := apiclientautenticacion.CheckTokenValido(w, r)
	if tokenValido {

		params := mux.Vars(r)
		novedad_id := params["id"]

		var novedad structNovedad.Novedad //Con &var --> lo que devuelve el metodo se le asigna a la var

		versionMicroservicio := obtenerVersionNovedad()
		tenant := apiclientautenticacion.ObtenerTenant(tokenAutenticacion)

		db := apiclientconexionbd.ObtenerDB(tenant, nombreMicroservicio, versionMicroservicio, AutomigrateTablasPrivadas)

		defer db.Close()

		//gorm:auto_preload se usa para que complete todos los struct con su informacion
		if err := db.Set("gorm:auto_preload", true).First(&novedad, "id = ?", novedad_id).Error; gorm.IsRecordNotFoundError(err) {
			framework.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		concepto := obtenerConcepto(*novedad.Conceptoid, r)
		novedad.Concepto = concepto
		framework.RespondJSON(w, http.StatusOK, novedad)
	}

}

func obtenerConcepto(conceptoid int, r *http.Request) *structConcepto.Concepto {

	var concepto structConcepto.Concepto

	config := configuracion.GetInstance()

	url := configuracion.GetUrlMicroservicio(config.Puertomicroservicioconcepto) + "concepto/conceptos/" + strconv.Itoa(conceptoid)

	//url := "http://localhost:8084/conceptos/" + strconv.Itoa(conceptoid)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	header := r.Header.Get("Authorization")

	req.Header.Add("Authorization", header)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Println("URL:", url)

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	str := string(body)

	json.Unmarshal([]byte(str), &concepto)

	return &concepto

}

func NovedadAdd(w http.ResponseWriter, r *http.Request) {

	tokenValido, tokenAutenticacion := apiclientautenticacion.CheckTokenValido(w, r)
	if tokenValido {

		decoder := json.NewDecoder(r.Body)

		var novedad_data structNovedad.Novedad

		if err := decoder.Decode(&novedad_data); err != nil {
			framework.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}

		defer r.Body.Close()

		versionMicroservicio := obtenerVersionNovedad()
		tenant := apiclientautenticacion.ObtenerTenant(tokenAutenticacion)

		db := apiclientconexionbd.ObtenerDB(tenant, nombreMicroservicio, versionMicroservicio, AutomigrateTablasPrivadas)

		defer db.Close()

		if err := db.Create(&novedad_data).Error; err != nil {
			framework.RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		framework.RespondJSON(w, http.StatusCreated, novedad_data)
	}
}

func NovedadUpdate(w http.ResponseWriter, r *http.Request) {

	tokenValido, tokenAutenticacion := apiclientautenticacion.CheckTokenValido(w, r)
	if tokenValido {

		params := mux.Vars(r)
		//se convirtió el string en int para poder comparar
		param_novedadid, _ := strconv.ParseInt(params["id"], 10, 64)
		p_novedadid := int(param_novedadid)

		if p_novedadid == 0 {
			framework.RespondError(w, http.StatusNotFound, framework.IdParametroVacio)
			return
		}

		decoder := json.NewDecoder(r.Body)

		var novedad_data structNovedad.Novedad

		if err := decoder.Decode(&novedad_data); err != nil {
			framework.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		defer r.Body.Close()

		novedadid := novedad_data.ID

		if p_novedadid == novedadid || novedadid == 0 {

			novedad_data.ID = p_novedadid

			versionMicroservicio := obtenerVersionNovedad()
			tenant := apiclientautenticacion.ObtenerTenant(tokenAutenticacion)

			db := apiclientconexionbd.ObtenerDB(tenant, nombreMicroservicio, versionMicroservicio, AutomigrateTablasPrivadas)

			defer db.Close()

			if err := db.Save(&novedad_data).Error; err != nil {
				framework.RespondError(w, http.StatusInternalServerError, err.Error())
				return
			}

			framework.RespondJSON(w, http.StatusOK, novedad_data)

		} else {
			framework.RespondError(w, http.StatusNotFound, framework.IdParametroDistintoStruct)
			return
		}
	}

}

func NovedadRemove(w http.ResponseWriter, r *http.Request) {

	tokenValido, tokenAutenticacion := apiclientautenticacion.CheckTokenValido(w, r)
	if tokenValido {

		//Para obtener los parametros por la url
		params := mux.Vars(r)
		novedad_id := params["id"]

		versionMicroservicio := obtenerVersionNovedad()
		tenant := apiclientautenticacion.ObtenerTenant(tokenAutenticacion)

		db := apiclientconexionbd.ObtenerDB(tenant, nombreMicroservicio, versionMicroservicio, AutomigrateTablasPrivadas)

		defer db.Close()

		//--Borrado Fisico
		if err := db.Unscoped().Where("id = ?", novedad_id).Delete(structNovedad.Novedad{}).Error; err != nil {

			framework.RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		framework.RespondJSON(w, http.StatusOK, framework.Novedad+novedad_id+framework.MicroservicioEliminado)
	}

}

func AutomigrateTablasPrivadas(db *gorm.DB) {

	//para actualizar tablas...agrega columnas e indices, pero no elimina
	db.AutoMigrate(&structNovedad.Novedad{})

}

func obtenerVersionNovedad() int {
	configuracion := configuracion.GetInstance()

	return configuracion.Versionnovedad
}
