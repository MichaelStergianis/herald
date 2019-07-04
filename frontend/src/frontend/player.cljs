(ns frontend.player
  (:require [clojure.core.async :as async :refer [chan go <! >! close!]]
            [ajax.core :as ajax :refer [GET]]
            [frontend.data :as data]
            [frontend.requests :as req]))

(defn play-song! [id]
  (GET (str "/" req/communication-protocol "/song/" id)
       {:handler (req/play-song-handler)}))
