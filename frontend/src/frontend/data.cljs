(ns frontend.data
  (:require [reagent.core :as r]))

(defonce viewport-dims (r/atom []))
(defonce active (r/atom :random))
(defonce libraries (r/atom (vector)))
(defonce artists (r/atom (vector)))
(defonce albums (r/atom (vector)))
(defonce sidebar-open (r/atom false))
(defonce sidebar-toggle-function (r/atom {:function 'toggle}))
(defonce manage-library (r/atom {}))

(defonce categories [{:name "Random" :class "la la-random" :set-active :random}
                     {:name "Artists" :class "la la-user" :set-active :artists}
                     {:name "Albums" :class "la la-folder-open" :set-active :albums}
                     {:name 'divider}
                     {:name "Settings" :class "la la-cog" :set-active :settings}])

(defonce settings [{:name "Manage Libraries" :set-active :manage-lib}])

(defonce player (r/atom {:visible false
                         :playing false}))
