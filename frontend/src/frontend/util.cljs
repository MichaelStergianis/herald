(ns frontend.util)

(defn by-id [id]
  (. js/document (getElementById id)))


(defn current-url []
  (.-href (.-location js/window)))
