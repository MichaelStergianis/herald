(ns frontend.ui
  (:require [frontend.util     :as util :refer [by-id]]
            [frontend.requests :as req]
            [frontend.state    :as state]))

(defn click-handler []
  ())

(defn artists []
  (req/get-all-artists)
  (fn []
    [:div#artists
     [:p "hello world"]]))

(defn navbar
  "The navbar"
  []
  (fn []
    [:nav {:class "navbar navbar-herald"}
     ;; logo
     [:a [:h2 "Herald"]]]))

(defn base []
  [:div
   [navbar]
   [artists]
   [:div [:i {:class "fas fa-play" :on-click click-handler}]]
   [:audio {:id "player-html5" }]])
