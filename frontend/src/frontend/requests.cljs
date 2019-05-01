(ns frontend.requests
  (:require [reagent.core :as r]
            [ajax.core :refer [GET POST]]))

(defonce communication-protocol "edn")

(defn write-to-state-handler! [response]
  nil)

(defn query-unique [table id]
  (GET (str "/" communication-protocol "/" table "/" id)))

(defn query [table data]
  (GET (str "/" communication-protocol "/")))

(defn get-all-artists []
  (GET (str "/" communication-protocol "/artists/")))
