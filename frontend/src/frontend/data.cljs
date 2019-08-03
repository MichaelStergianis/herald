(ns frontend.data
  (:require [reagent.core :as r]))

(defonce viewport-dims (r/atom []))
(defonce scroll-position (r/atom 0))

(defonce active (r/atom :random))
(defonce libraries (r/atom (vector)))
(defonce songs (r/atom (vector)))
(defonce artists (r/atom (vector)))
(defonce albums (r/atom (vector)))
(defonce sidebar-open (r/atom false))
(defonce sidebar-toggle-function (r/atom {:function 'toggle}))
(defonce manage-library (r/atom {}))

(defonce categories [{:name "Random" :class "la la-random" :set-active :random}
                     {:name "Songs" :class "la la-file-sound-o" :set-active :songs}
                     {:name "Artists" :class "la la-user" :set-active :artists}
                     {:name "Albums" :class "la la-folder-open" :set-active :albums}
                     {:name 'divider}
                     {:name "Settings" :class "la la-cog" :set-active :settings}])

(defonce settings [{:name "Manage Libraries" :set-active :manage-lib}])

(defonce player (r/atom {:playing false
                         :paused true}))
(defonce volume (r/atom 1))

(defonce space-char 32)

(defonce player-html (r/atom nil))
(defonce audio-html (r/atom nil))
(defonce player-props (r/atom {}))
