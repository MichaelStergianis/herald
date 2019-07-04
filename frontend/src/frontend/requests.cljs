(ns frontend.requests
  (:require [cljs.reader]
            [frontend.data :as data]
            [reagent.core :as r]
            [ajax.core :refer [GET POST PUT DELETE]]))

(def parser cljs.reader/read-string)
(def communication-protocol "edn")

(defn req-str [table]
  (str "/" communication-protocol "/" table))

(defn reset-handler [data-loc data-trans-fn]
  (fn [response]
    (let [appended (data-trans-fn (parser response))]
      (reset! data-loc appended))))

(defn assoc-in-handler [data-trans-fn data-loc tags]
  (fn [response]
    (let [appended (data-trans-fn (parser response))]
      (swap! data-loc assoc-in tags appended))))

(defn album-handler [album-data artist-data]
  (fn [response]
    (let [album-resp (parser response)]
      (reset! album-data album-resp)
      (when (@album-data :artist)
        (GET (str "/" communication-protocol "/" "artist/" (@album-data :artist))
             {:handler (reset-handler artist-data (fn [response] response))})))))

(defn play-song-handler []
  (fn [response]
    (swap! data/player assoc :playing true)
    (swap! data/player assoc :paused false)
    (let [song-resp (parser response)]
      (swap! data/player assoc :song song-resp)
      (when (get-in @data/player [:song :album])
        (GET (str "/" communication-protocol "/album/" (get-in @data/player [:song :album]))
             {:handler (fn [response]
                         ((assoc-in-handler identity data/player [:album]) response)
                         (GET (str "/" communication-protocol "/artist/" (get-in @data/player [:album :artist]))
                              {:handler (assoc-in-handler identity data/player [:artist])}))})))))

(defn query-unique [table id]
  (GET (str "/" communication-protocol "/" table "/" id)))

(defn query [table data]
  (GET (str "/" communication-protocol "/")))

(defn get-artists []
  (GET (str "/" communication-protocol "/artists/")
       {:params {:orderby "id"}
        :handler (reset-handler data/artists #(%1 0))}))

(defn get-all
  "Gets all of a given media type by "
  [media data-loc & order-by]
  (GET (str "/" communication-protocol "/" media)
       {:params {:orderby order-by}
        :handler (reset-handler data-loc #(%1 0))}))

(defn get-unique
  "Gets a single type as specified"
  [id media data-loc]
  (GET (str "/" communication-protocol "/" media "/" id)
       {:handler (reset-handler data-loc (fn [i] i))}))


(defn scan-library
  "POSTS a request to the server to re-scan the given library."
  [id]
  (POST (str "/" communication-protocol "/scanLibrary/" id)
        {:handler #(println %)}))

(defn create-library [name path]
  (POST (str "/" communication-protocol "/library")
      {:body (str {:name name :path path})
       :handler #(swap! data/libraries conj (parser %))}))

(defn update-library [id name path]
  (println id name path)
  (PUT (str "/" communication-protocol "/library")
       {:body (str {:id id :name name :path path})}))

(defn get-song [id & {:keys [from to]}]
  (GET (str "/" communication-protocol "/stream/" id)
       {:headers (if (every? (comp nil? not) '(from to)) {:range (str "bytes: " from "-" to)} {})
        :handler #(.log js/console %)}))
