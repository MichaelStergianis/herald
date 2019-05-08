(ns ^:figwheel-hooks frontend.core
  (:require [cljss.core :as css]
            [frontend.ui   :refer [base]]
            [frontend.data :as data]
            [frontend.styles :as s]
            [frontend.util :as util :refer [by-id]]
            [reagent.core  :as r]))

(enable-console-print!)

(defn render! []
  (r/render-component [base] (by-id "app")))

(defonce on-create
  (do
    (let [viewport-fn (fn [] (swap! data/viewport-dims conj
                                   [:width (.-innerWidth js/window)]
                                   [:height (.-innerHeight js/window)]))]
      (viewport-fn)
      (.addEventListener js/window "resize" viewport-fn))

    (.appendChild (.-head js/document)
                  (let [meta-elem (.createElement js/document "meta")]
                    (.setAttribute meta-elem "name" "theme-color")
                    (.setAttribute meta-elem "content" s/p-dark)
                    meta-elem))

    (.setAttribute (.-body js/document) "style" "margin: 0px;")
    true))

(defn -main []
    (render!))

(defn ^:after-load on-js-reload []
  (css/remove-styles!)
  (-main))

(do
  (-main))
