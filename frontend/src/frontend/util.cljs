(ns frontend.util
  (:require [goog.string :as gstring]
            [goog.string.format]))

(defn by-id [id]
  (. js/document (getElementById id)))

(defn current-url []
  (.-href (.-location js/window)))

(defn ref-handler [state]
  (fn [elem] (when elem (reset! state elem))))

(defn format-time [t d]
  (let [h (/ t 3600)
        m (/ t 60)
        s (rem t 60)
        dh (/ d 3600)
        dm (/ d 60)]
    (str
     (if (> dh 1) (gstring/format "%02d:" h))
     (if (> dm 1) (gstring/format "%02d:" m) "00:")
     (gstring/format "%02d" s))))
