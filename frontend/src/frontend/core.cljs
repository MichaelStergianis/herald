(ns ^:figwheel-hooks frontend.core
  (:require [cljss.core :as css]
            [frontend.ui   :refer [base]]
            [frontend.util :as util :refer [by-id]]
            [reagent.core  :as r    :refer [atom]]))

(enable-console-print!)

(defn render! []
  (r/render-component [base] (by-id "app")))

(defn -main []
  (.setAttribute (.-body js/document) "style" "margin: 0px;")
  (render!))

(defn ^:after-load on-js-reload []
  (css/remove-styles!)
  (-main))

(-main)
