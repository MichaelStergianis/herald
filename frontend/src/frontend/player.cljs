(ns frontend.player
  (:require [clojure.core.async :as async :refer [chan go <! >! close!]]
            [reagent.core :as r]
            [ajax.core :as ajax :refer [GET]]
            [frontend.styles :as s :refer [compose]]
            [frontend.data :as data]
            [frontend.util :as u]
            [frontend.requests :as req]))

(declare audio progress-bar handle media-timing media-controls
         media-metadata volume-input play-song! play-song-handler
         play-pause! seek-track!)

(defn playlist-metadata []
  ())

(defn player [state]
  (let [player-id (gensym "player_")]
    (r/create-class
     {:component-did-mount
      (fn []
        (if (.-mediaSession js/navigator)
          (let [song (@data/player :song)
                album (@data/player :album)]
            (set! (-> js/navigator .-mediaSession .-metadata)
                  (new js/MediaMetadata
                       #js{:title (song :title)
                           :artist (song :artist)}))
            (doall (for [handler '("play" "pause")]
                     (-> js/navigator .-mediaSession (.setActionHandler handler (play-pause! state))))))))
      :reagent-render
      (fn [state]
        (when @data/audio-html
          (if (@state :paused)
            (.pause @data/audio-html)
            (.play @data/audio-html)))
        [:div {:class (s/player)
               :ref (u/ref-handler data/player-html)}
         [handle state]
         [media-timing state]
         [:div {:class (compose (s/player-bottom-area))}
          [media-controls state]
          [media-metadata state]
          [volume-input state]]
         [audio state player-id]])})))

(defn audio [state player-id]
  (let [this (r/current-component)
        ticking? (atom false)
        rerender? (r/atom false)
        refresh-interval 333 #_"milliseconds"] 
    (fn [state player-id]
      (when-let [playlist-idx (get-in @state [:playlist :playlist-idx])]
        (let [song-id (get-in @state [:song :id])]
          [:audio {:id player-id :volume @data/volume
                   :on-ended (partial seek-track! state inc)
                   :on-play (fn []
                              (when (not @ticking?)
                                (reset! ticking? true)
                                (set! (.-timeRefresher this)
                                      (.setInterval
                                       js/window
                                       (fn []
                                         (let [audio @data/audio-html
                                               play-position (.-currentTime audio)
                                               duration      (.-duration audio)
                                               buffered      (.-buffered audio)]
                                           (reset! data/player-props
                                                   {:play-position play-position
                                                    :duration      duration
                                                    :buffered      buffered})))
                                       refresh-interval))))
                   :on-pause (fn []
                               (when @ticking?
                                 (reset! ticking? false)
                                 ;; (swap! state assoc :paused true)
                                 (.clearInterval js/window (.-timeRefresher this))))
                   :ref (fn [elem] (when elem (reset! data/audio-html elem)))
                 :autoPlay (not (@state :paused))
                 :src (str "/edn/stream/" song-id)}])))))

(defn handle [state]
  (let [active? (r/atom false)]
   (fn [state]
    (let [[w _] @data/viewport-dims]
      [:div {:class (s/player-handle-area)}
       [:div {:class (compose (s/player-handle) (s/margin-botom 8))
              :on-click (fn [] (swap! active? not))}]
       [:div {:class (compose (s/player-handle-info) (when @active? (s/player-handle-info-active)))}
        [:div {:class (compose (s/fg "#808080") (s/player-handle-elem w))}
         [:div ]
         [:div {:class (compose (s/display "inline"))} "Title"]
         [:div {:class (compose (s/display "inline"))} "Album"]
         [:div {:class (compose (s/display "inline"))} "Artist"]
         [:div {:class (compose (s/display "inline"))} "Duration"]]
        (if-let [playlist (@state :playlist)]
          (doall
           (for [t (playlist :track-order)]
             (let [s ((playlist :songs) t)
                   albums (playlist :albums)
                   a (if albums (albums (s :album)))]
               [:div {:key (s :track) :class (compose (s/player-handle-elem w))}
                [:div]
                [:div {:class (compose (s/display "inline"))} (s :title)]
                [:div {:class (compose (s/display "inline"))} (if a (a :title) "Unknown Album")] 
                [:div {:class (compose (s/display "inline"))} (s :artist)]
                [:div {:class (compose (s/display "inline"))} (u/format-time (s :duration) (s :duration))]]))))]]))))

(defn progress-bar [state play-position duration]
  (let [seek-area (r/atom nil)]
    (fn [state play-position duration]
      [:div {:class (compose (s/player-slider-area) (s/no-select))
             :ref (u/ref-handler seek-area)
             :on-click (fn [e] (if-let [sa @seek-area]
                                (let [w (.-offsetWidth sa)
                                      l (.-offsetLeft sa)
                                      click-loc (-> e .-clientX)
                                      percent (/ (- click-loc l) w)
                                      time (* percent duration)]
                                  (set! (-> @data/audio-html .-currentTime) time)
                                  (swap! state assoc :paused false))))}
       (when (not (empty? @data/player-props))
         (doall
          (for [i (range (.-length (@data/player-props :buffered)))]
            (let [buffer   (@data/player-props :buffered)
                  start    (.start buffer i)
                  end      (.end buffer i)
                  left     (* 100 (/ start duration))
                  width    (* 100 (/ (- end start) duration))]
              [:div {:key i :class (compose (s/buffered-slider))
                     :style {:left (str left "%") :width (str width "%")}}]))))
       [:div {:class (compose (s/player-slider))}]
       [:div {:class (compose (s/played-slider)) :style {:width (str (* 100 (/ play-position duration)) "%")}}]
       [:div {:class (compose (s/player-cursor)) :style {:left  (str (* 100 (/ play-position duration)) "%")}}]])))



(defn media-timing [state]
  (let [player-props @data/player-props
        play-position (player-props :play-position)
        duration      (player-props :duration)]
    [:div {:class (compose (s/playing-stats))}
     [:div {:class (compose (s/player-slider-time))} (u/format-time play-position duration)]
     [progress-bar state play-position duration]
     [:div {:class (compose (s/player-slider-time))} (u/format-time duration duration)]]))

(defn media-controls [state]
  [:div {:class (compose (s/player-control-area))}
   [:button {:title "Previous"
             :class (compose (s/no-select) (s/circle-bounding) (s/player-button) "la la-fast-backward")
             :on-click (partial seek-track! state dec)}]
   [:button {:title "Play"
             :class (compose (s/no-select) (s/circle-bounding)
                             (s/player-button) (s/player-play-button) "la"
                             (if (@state :paused) "la-play" "la-pause"))
             :on-click (play-pause! state)
             :on-key-press (fn [e]
                             (when (and (= data/space-char (-> e .-charCode))
                                        (.-stopPropagation e))
                               (.stopPropagation e)))}]
   [:button {:title "Next"
             :class (compose (s/no-select) (s/circle-bounding)
                             (s/player-button) "la la-fast-forward")
             :on-click (partial seek-track! state inc)}]])

(defn media-metadata [state]
  [:div {:class (compose (s/player-info-area))}
   [:div (get-in @state [:song :title])]
   [:div (get-in @state [:album :title])]
   [:div (get-in @state [:artist :name])]])

(defn volume-input [state]
  [:div {:class (compose (s/right) (s/player-volume-area))}
   [:div {:class (compose (s/margin "0 4px") (s/display "inline")) } (int (* 100 @data/volume))]
   [:input {:type "range" :min "0" :max "1" :step "0.001" :value @data/volume
            :class (compose  (s/player-volume-slider 6))
            :on-change (fn [e] (let [v (-> e .-target .-value)]
                                (when @data/audio-html
                                  (set! (-> @data/audio-html .-volume) v))
                                (reset! data/volume v)))}]])

(defn play-playlist!
  "Constructs and begins playing a playlist"
  [metadata]
  (swap! data/player assoc :playlist metadata)
  (GET (str "/" req/communication-protocol "/song/" (nth (metadata :track-order) (metadata :playlist-idx)))
       {:handler play-song-handler}))

(defn seek-track! [state f]
  (let [new-idx (f (get-in @state [:playlist :playlist-idx]))]
    (when (and (<= 0 new-idx) (> (count (get-in @state [:playlist :track-order])) new-idx))
      (swap! state update-in [:playlist :playlist-idx] f)
      (swap! state assoc :song (get-in @state [:playlist :songs (get-in @state [:playlist :track-order new-idx])]))
      (swap! state assoc :album (get-in @state [:playlist :albums (get-in @state [:song :album])])))))

(defn play-song-handler [response]
  (swap! data/player assoc :playing true)
  (swap! data/player assoc :paused false)
  (let [song-resp (req/parser response)]
    (swap! data/player assoc :song song-resp)
    (when (get-in @data/player [:song :album])
      (GET (str "/" req/communication-protocol "/album/" (get-in @data/player [:song :album]))
           {:handler (fn [response]
                       ((req/assoc-in-handler identity data/player [:album]) response)
                       (when (get-in @data/player [:album :artist])
                         (GET (str "/" req/communication-protocol "/artist/" (get-in @data/player [:album :artist]))
                              {:handler (req/assoc-in-handler identity data/player [:artist])})))}))))

(defn play-pause! [state]
  (fn [] (swap! state update :paused not)))
