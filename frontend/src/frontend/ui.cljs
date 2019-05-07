(ns frontend.ui
  (:require [reagent.core :as r]
            [clojure.string :refer [lower-case]]
            [cljss.reagent :refer-macros [defstyled]]
            [frontend.styles :as s :refer [compose]]
            [frontend.util     :as util :refer [by-id]]
            [frontend.requests :as req]
            [frontend.data    :as data]))

(defn set-active! [k]
  (reset! data/active k))

(defonce navbar-height 48)
(defonce medium-bar-divisor 4)
(defonce small-bar-divisor 24)
(defonce total-navbar-height (+ navbar-height
                                (/ navbar-height medium-bar-divisor)
                                (/ navbar-height small-bar-divisor)))
(defonce sidebar-width 160)

(defstyled padded-div :div
  {:margin-top (str total-navbar-height "px")
   :padding "16px"
   :height "100%"})

(defn random []
  (let [get-random-songs
        (fn []
          )]
    (fn []
      [padded-div "Random"])))

(defn artists []
  (req/get-all "artists" data/artists "name")
  (fn []
    [padded-div {:id "artists"}
     (for [artist @data/artists]
       [:div {:key (str "artist-" (artist :id))} (artist :name)])]))

(defn albums []
  (req/get-all "albums" data/albums "title")
  (fn []
    [padded-div {:id "albums"}
     (for [album @data/albums]
       [:div {:key (str "album-" (album :id))} (album :title)])]))

(defn toggle-sidebar-visibility! []
  (reset! data/sidebar-open (not @data/sidebar-open)))

(defn sidebar-toggle []
  (fn []
    [:button {:class (compose (s/navbar-toggle navbar-height))
              :on-click toggle-sidebar-visibility!}
     [:span {:class (s/sr-only)} "Toggle Navigation"]
     [:i {:class (compose (if @data/sidebar-open (s/color-on-active s/secondary)) (s/circle-bounding) "la la-bars")}]]))

(defn options-toggle []
  (let [button-active (r/atom false)]
    (fn []
      [:button {:class (compose (s/navbar-toggle navbar-height) (s/right))
                :on-click (fn [] (reset! button-active (not @button-active)))}
       [:span {:class (s/sr-only)} "Options"]
       [:i {:class (compose (if @button-active (s/color-on-active s/secondary)) (s/circle-bounding) "la la-ellipsis-v")}]])))

(defn sidebar-li-click-event! [keyw]
  (fn [e]
    (reset! data/sidebar-open false)
    (set-active! keyw)))

(defn sidebar []
  (fn []
    [:div {:class (compose
                   ;; subtracting 1 from the innerheight prevents
                   ;; the sidebar from exceeding the viewport size
                   (s/sidebar (- (.-innerHeight js/window) 1) total-navbar-height sidebar-width)
                   (if @data/sidebar-open (s/sidebar-open)))}
     [:ul {:class (s/sidebar-ul)}
      (doall (for [item [{:name "Random" :class "la la-random"}
                         {:name "Artists" :class "la la-user"}
                         {:name "Albums" :class "la la-folder-open"}]]
               (let [keyw (keyword (lower-case (item :name)))]
                 [:li {:key (item :name)
                       :class (compose
                               (s/sidebar-li 5)
                               (if (= @data/active keyw) (s/sidebar-li-active)))
                       :on-click (sidebar-li-click-event! keyw)}
                  [:a {:class (compose
                               (item :class)
                               (s/sidebar-li-icon))}]
                  [:a {:class (compose (s/sidebar-li-a)
                                       (s/pad-in-start 8)
                                       (s/right))
                       :style {:padding-right 5}} (item :name)]])))]]))

(defn navbar
  "Creates a navigation bar"
  []
  (fn []
    [:div#nav-area {:class (s/navbar)}
     [:div {:class (s/above-nav (int (/ navbar-height medium-bar-divisor)))}]
     [:div {:class (s/between-above-nav (int (/ navbar-height small-bar-divisor)))}]
     [:div {:class (compose (s/pad-in-start 5) (s/navbar-nav navbar-height))}
      ;; toggle
      [sidebar-toggle]
      ;; logo
      [:a.navbar-brand
       {:class (compose
                (s/pad-in-start 10)
                (s/navbar-brand))
        :on-click #((toggle-sidebar-visibility!) (set-active! :random))} "herald"]
      ;; options
      [options-toggle]]]))

(defn player []
  [:audio {:id "player-html5"}])

(defn base []
  [:div {:class (s/roboto-font)}
   [navbar]
   [sidebar]
   (case @data/active
     :random [random]
     :artists [artists]
     :albums [albums]
     ;; default
     [random])
   [player]])

