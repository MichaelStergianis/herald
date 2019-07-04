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
    (let [viewport-fn (fn [] (reset! data/viewport-dims
                                    [(.-innerWidth js/window) (.-innerHeight js/window)]))]
      (viewport-fn)
      (.addEventListener js/window "resize" viewport-fn))

    (.addEventListener js/window "keypress"
                       (fn [e]
                         (let [k (-> e .-charCode)]
                           (when (and (= k 32) (@data/player :playing) (.-preventDefault e))
                             (.preventDefault e)
                             (swap! data/player update :paused not)))))

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
