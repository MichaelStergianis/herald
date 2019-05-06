(ns frontend.requests
  (:require [cljs.reader]
            [frontend.data :as data]
            [reagent.core :as r]
            [ajax.core :refer [GET POST]]))

(def parser cljs.reader/read-string)
(def communication-protocol "edn")

(defn create-write-to-data-handler [data-loc data-trans-fn]
  (fn [response]
    (let [appended (data-trans-fn (parser response))]
      (reset! data-loc appended))))

(defn query-unique [table id]
  (GET (str "/" communication-protocol "/" table "/" id)))

(defn query [table data]
  (GET (str "/" communication-protocol "/")))

(defn get-artists []
  (GET (str "/" communication-protocol "/artists/")
       {:params {:orderby "id"}
        :handler (create-write-to-data-handler data/artists #(%1 0))}))

(defn get-all
  "Gets all of a given media type by "
  [media data-loc & order-by]
  (GET (str "/" communication-protocol "/" media "/")
       {:params {:orderby order-by}
        :handler (create-write-to-data-handler data-loc #(%1 0))}))
