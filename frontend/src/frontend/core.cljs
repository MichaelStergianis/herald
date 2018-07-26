(ns frontend.core
  (:require [frontend.util :as util :refer [by-id]]
            [reagent.core  :as r    :refer [atom]]))

(enable-console-print!)

(defn click-handler []
  ())

(defn base []
  [:div
   [:div [:p "Hello world"]]
   [:div [:i {:class "fas fa-play" :on-click click-handler}]]
   [:audio {:id "player-html5" }]
   ])

(defn render! []
  (r/render-component [base]
                      (by-id "app")))

(defn -main []
  (render!))

(defn on-js-reload []
  (-main))

(-main)
