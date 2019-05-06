(ns frontend.data
  (:require [reagent.core :as r]))

(defonce active (r/atom :random))
(defonce artists (r/atom (vector)))
(defonce albums (r/atom (vector)))
(defonce sidebar-open (r/atom false))

