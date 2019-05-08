(ns ^:figwheel-hooks frontend.core
  (:require [cljss.core :as css]
            [frontend.ui   :refer [base]]
            [frontend.data :as data]
            [frontend.util :as util :refer [by-id]]
            [reagent.core  :as r    :refer [atom]]))

(enable-console-print!)

(defn render! []
  (r/render-component [base] (by-id "app")))

(defn -main []
  (let [viewport-fn (fn [] (swap! data/viewport-dims conj
                                 [:width (.-innerWidth js/window)]
                                 [:height (.-innerHeight js/window)]))]
    (.setAttribute (.-body js/document) "style" "margin: 0px;")
    (render!)
    (viewport-fn)
    (.addEventListener js/window "resize" viewport-fn)))

(defn ^:after-load on-js-reload []
  (css/remove-styles!)
  (-main))

(-main)
