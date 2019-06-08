(ns frontend.ui
  (:require [reagent.core :as r]
            [reagent.dom :refer [dom-node]]
            [ajax.core :refer [GET]]
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


(defn full-screen-backdrop
  "Provides a full screen exit button for whatever context"
  [state z-index]
  (fn [state z-index]
    [:div {:class (compose
                   (let [[w h] @data/viewport-dims] (s/full-screen-backdrop w h z-index))
                   (if @state (s/full-screen-backdrop-active)))}]))

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

(defn album-button []
  (let [this (r/current-component)
        clicked? (r/atom false)]
    (fn []
      [:i (r/merge-props {:class (compose (s/circle-bounding) (s/album-button 4 4) (if @clicked? (s/album-button-clicked)))
                          :on-click (fn [] (reset! clicked? (not @clicked?)))}
                         (r/props this))])))

(defn album []
  (let [album-info  (r/atom {})
        artist-info (r/atom {})
        mouse-on?    (r/atom false)
        this (r/current-component)
        width-height 168]
    (GET (str (req/req-str "album") "/" ((r/props this) :albumid))
         {:handler (req/album-handler album-info artist-info)})
    (fn []
      [:div (r/merge-props {:class (s/album width-height 2)
                            :on-mouse-enter (fn [] (reset! mouse-on? true))
                            :on-mouse-leave (fn [] (reset! mouse-on? false))
                            :on-click (fn [] (println "album" ((r/props this) :albumid) "clicked"))}
                           (r/props this))
       [:div {:class (s/album-inside)}
        [:div {:class (s/album-background)}]
        [:i {:class (compose "la la-music" (s/album-img))}]
        [:div {:class (compose (s/no-select) (s/album-info))}
         [:b (@artist-info :name)]
         [:br]
         (@album-info :title)]
        [:div {:class (compose (s/album-buttons) (if @mouse-on? (s/album-buttons-show)))}
         [album-button {:class (compose "la la-bookmark")}]
         [album-button {:class (compose "la la-play")}]]]])))

(defn albums []
  (req/get-all "albums" data/albums "title")
  (fn []
    [padded-div {:id "albums"}
     (for [a @data/albums]
       [album {:key (str "album-" (a :id))
               :albumid (a :id)}])]))

(defn toggle-visibility! [state]
  (swap! state not))

(defn sidebar-toggle []
  (fn []
    [:button {:class (compose (s/navbar-toggle navbar-height))
              :on-click #(toggle-visibility! data/sidebar-open)}
     [full-screen-backdrop data/sidebar-open "998"]
     [:span {:class (s/sr-only)} "Toggle Navigation"]
     [:i {:class (compose (if @data/sidebar-open (s/color-on-active s/secondary))
                          (s/toggle) (s/circle-bounding) "la la-bars")}]]))

(defn sidebar-li-click-event! [keyw]
  (fn [e]
    (reset! data/sidebar-open false)
    (set-active! keyw)))

(defn sidebar []
  (fn []
    [:div {:class (compose
                   ;; subtracting 1 from the innerheight prevents
                   ;; the sidebar from exceeding the viewport size
                   (let [[_ height] @data/viewport-dims]
                     (s/sidebar (- height 1) total-navbar-height sidebar-width))
                   (if @data/sidebar-open (s/sidebar-open)))}
     [:ul {:class (s/sidebar-ul)}
      (doall (for [item data/categories]
               (let [keyw (keyword (lower-case (item :name)))]
                 [:li {:key (item :name)
                       :class (compose
                               (s/menu-li)
                               (if (= @data/active keyw) (s/highlighted-row)))
                       :on-click (sidebar-li-click-event! keyw)}
                  [:a {:class (compose
                               (item :class)
                               (s/sidebar-li-icon))}]
                  [:a {:class (compose (s/sidebar-li-a)
                                       (s/pad-in-start 8)
                                       (s/right))
                       :style {:padding-right 5}} (item :name)]])))]]))

(defn options-menu [button-active op-toggle]
  (fn [button-active op-toggle]
    (when (-> @op-toggle
             nil?
             not)
      [:div
       [full-screen-backdrop button-active 98]
       (let [top   (.-offsetHeight @op-toggle) 
             right (.-offsetWidth @op-toggle)]
         [:div {:class (compose (s/options-menu top right) (if @button-active (s/options-menu-active)))}
          (doall (for [elem [{:content "Manage Libraries" :key "manage-libs"
                              :click (fn [] (println "hello"))}]]
                   [:div {:key (str "options-" (elem :key))
                          :class (compose (s/menu-li))
                          :on-click (elem :click)}
                    (elem :content)]))])])))

(defn options-toggle []
  (let [button-active (r/atom false)
        op-toggle (r/atom nil)]
    (fn []
      [:button {:id "options-toggle"
                :ref #(reset! op-toggle %)
                :class (compose (s/navbar-toggle navbar-height) (s/right))
                :on-click #(toggle-visibility! button-active)}
       [:span {:class (s/sr-only)} "Options"]
       [:i {:class (compose (if @button-active (s/color-on-active s/secondary))
                            (s/toggle) (s/circle-bounding) "la la-ellipsis-v")}]
       [options-menu button-active op-toggle]])))

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
        :on-click #((reset! data/sidebar-open false) (set-active! :random))} "Warbler"]
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

