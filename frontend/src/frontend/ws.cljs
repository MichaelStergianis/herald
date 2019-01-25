(ns frontend.ws
  (:require [frontend.util :refer [current-url]]
            [clojure.string :refer [replace]]))

(defn make-ws [url]
  (new js/WebSocket url))

(defn print! [message]
  (.log js/console message))
