(ns frontend.core
  (:require [frontend.ui   :refer [base]]
            [frontend.util :as util :refer [by-id]]
            [reagent.core  :as r    :refer [atom]]))

(enable-console-print!)

(defn render! []
  (r/render-component [base] (by-id "app")))

(defn -main []
  (render!))

(defn on-js-reload []
  (-main))

(-main)
