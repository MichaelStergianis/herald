(ns frontend.ui
  (:require [reagent.core :as r]
            [clojure.string :refer [lower-case]]
            [cljss.reagent :refer-macros [defstyled]]
            [frontend.styles :as styles :refer [compose]]
            [frontend.util     :as util :refer [by-id]]
            [frontend.requests :as req]
            [frontend.state    :as state]))

(defn set-active! [k]
  (reset! state/active k))

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
  [padded-div "Random"])

(defn artists []
  (req/get-all "artists" state/artists "name")
  (fn []
    [padded-div {:id "artists"}
     (for [artist @state/artists]
       [:div {:key (str "artist-" (artist :id))} (artist :name)])]))

(defn albums []
  (req/get-all "albums" state/albums "title")
  (fn []
    [padded-div {:id "albums"}
     (for [album @state/albums]
       [:div {:key (str "album-" (album :id))} (album :title)])]))

(defn navbar
  "Creates a navigation bar"
  []
  (fn []
    [:div#nav-area {:class (styles/navbar)}
     [:div {:class (styles/above-nav (int (/ navbar-height medium-bar-divisor)))}]
     [:div {:class (styles/between-above-nav (int (/ navbar-height small-bar-divisor)))}]
     [:div {:class (compose (styles/pad-in-start 5) (styles/navbar-nav navbar-height))}
      ;; toggle
      [:button {:class (styles/navbar-toggle)
                :on-click #(reset! state/sidebar-open (not @state/sidebar-open))}
       [:span.sr-only "Toggle Navigation"]
       [:i {:class "fas fa-bars"}]]
      ;; logo
      [:a.navbar-brand
       {:class (compose
                (styles/pad-in-start 10)
                (styles/navbar-brand))
        :on-click #(set-active! :random)} "herald"]]]))

(defn sidebar-li-click-event! [this keyw]
  (fn [e]
    (reset! state/sidebar-open false)
    (set-active! keyw)))

(defn sidebar []
  (let [this (r/current-component)]
    (fn []
      [:div {:class (compose
                     (styles/sidebar (.-innerHeight js/window) total-navbar-height sidebar-width)
                     (if @state/sidebar-open (styles/sidebar-open)))}
       [:ul {:class (styles/sidebar-ul)}
        (doall (for [item [{:name "Artists" :class "fas fa-user"}
                           {:name "Albums" :class "fas fa-compact-disc"}]]
                 (let [keyw (keyword (lower-case (item :name)))]
                   [:li {:key (item :name)
                         :class (compose
                                 (styles/sidebar-li 5)
                                 (if (= @state/active keyw) (styles/sidebar-li-active)))
                         :on-click (sidebar-li-click-event! this keyw)}
                    [:a {:class (compose
                                 (item :class)
                                 (styles/sidebar-li-a))}]
                    [:a {:class (compose (styles/sidebar-li-a)
                                         (styles/pad-in-start 8)
                                         (styles/right))
                         :style {:padding-right 5}} (item :name)]])))]])))

(defn player []
  [:audio {:id "player-html5"}])

(defn base []
  [:div
   [navbar]
   [sidebar]
   (case @state/active
     :random [random]
     :artists [artists]
     :albums [albums]
     ;; default
     [random])
   [player]])

