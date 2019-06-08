(ns frontend.data
  (:require [reagent.core :as r]))

(defonce viewport-dims (r/atom []))
(defonce active (r/atom :random))
(defonce artists (r/atom (vector)))
(defonce albums (r/atom (vector)))
(defonce sidebar-open (r/atom false))

(defonce categories [{:name "Random" :class "la la-random"}
                     {:name "Artists" :class "la la-user"}
                     {:name "Albums" :class "la la-folder-open"}])
