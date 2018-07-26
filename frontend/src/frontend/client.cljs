(ns frontend.client
  (:require [ajax.core :refer [GET POST]]))

(defonce server-addr (.-host js/window.location))

(println "hello")

(defn uri->url [uri]
  (str (.-protocol js/window.location) "//" server-addr "/" uri))

(defn fetch-music []
  #_(GET ))
