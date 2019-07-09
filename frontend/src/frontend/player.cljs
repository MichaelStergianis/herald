(ns frontend.player
  (:require [clojure.core.async :as async :refer [chan go <! >! close!]]
            [reagent.core :as r]
            [ajax.core :as ajax :refer [GET]]
            [goog.string :as gstring]
            [goog.string.format]
            [frontend.styles :as s :refer [compose]]
            [frontend.data :as data]
            [frontend.requests :as req]))

(defn play-song! [id]
  (GET (str "/" req/communication-protocol "/song/" id)
       {:handler (req/play-song-handler)}))

(defn play-pause! [state]
  (fn [] (swap! state update :paused not)))

(defn audio [state audio-elem player-id]
  (fn [state audio-elem player-id]
    (if-let [song-id (get-in @state [:song :id])]
      (let [src (str "/edn/stream/" song-id)]
        (when @audio-elem
          (if (@state :paused)
            (.pause @audio-elem)
            (.play @audio-elem)))
        [:audio {:id player-id :volume @data/volume
                 :ref #(reset! audio-elem %)
                 :autoPlay (not (@state :paused)) :src src}]))))

(defn player [state]
  (let [player-id (gensym "player_")
        seek-area (r/atom nil)
        audio-elem (r/atom nil)]
    (fn [state]
      (if (.-mediaSession js/navigator)
        (let [song (@data/player :song)
              album (@data/player :album)]
          (set! (-> js/navigator .-mediaSession .-metadata)
                (new js/MediaMetadata
                     #js{:title (song :title)
                         :artist (song :artist)}))
          (doall (for [handler '("play" "pause")]
                   (-> js/navigator .-mediaSession (.setActionHandler handler (play-pause! state)))))))

      ;; handle
      [:div {:class (s/player)
             :ref #(reset! data/player-html %)}
       [:div {:class (s/player-handle-area)}
        [:div {:class (s/player-handle)
               :on-click (fn [] (println "handle clicked"))}]]

       ;; play position
       (let [audio @audio-elem
             play-position (if (nil? audio) 0 (.-currentTime audio))
             duration      (if (nil? audio) 100 (.-duration audio))
             buffered      (when (not (nil? audio)) (.-buffered audio))
             format-time (fn [t d]
                           (let [h (/ t 3600)
                                 m (/ t 60)
                                 s (rem t 60)
                                 dh (/ d 3600)
                                 dm (/ d 60)]
                             (str
                              (if (> dh 1) (gstring/format "%02d:" h))
                              (if (> dm 1) (gstring/format "%02d:" m) "00:")
                              (gstring/format "%02d" s))))]
         [:div {:class (compose (s/playing-stats))}
          [:div {:class (compose (s/player-slider-time))} (format-time play-position duration)]
          [:div {:class (compose (s/player-slider-area) (s/no-select))
                 :ref #(reset! seek-area %)
                 :on-click (fn [e] (if-let [sa @seek-area]
                                    (let [w (.-offsetWidth sa)
                                          l (.-offsetLeft sa)
                                          click-loc (-> e .-clientX)
                                          percent (/ (- click-loc l) w)
                                          time (* percent (.-duration audio))]
                                      (set! (-> audio .-currentTime) time))))}
           (when buffered
             (doall
              (for [i (range (-> audio .-buffered .-length))]
                (let [buffer   (-> audio .-buffered)
                      duration (-> audio .-duration)
                      start    (.start buffer i)
                      end      (.end buffer i)
                      left     (* 100 (/ start duration))
                      width    (* 100 (/ (- end start) duration))]
                  [:div {:key i :class (compose (s/buffered-slider))
                         :style {:left (str left "%") :width (str width "%")}}]))))
           [:div {:class (compose (s/player-slider))}]
           [:div {:class (compose (s/played-slider)) :style {:width (str (* 100 (/ play-position duration)) "%")}}]
           [:div {:class (compose (s/player-cursor)) :style {:left  (str (* 100 (/ play-position duration)) "%")}}]]
          [:div {:class (compose (s/player-slider-time))} (format-time duration duration)]])

       [:div {:class (compose (s/player-bottom-area))}

        ;; controls
        [:div {:class (compose (s/player-control-area))}
         [:button {:title "Previous" :class (compose (s/no-select) (s/circle-bounding)
                                                     (s/player-button) "la la-fast-backward")
                   :on-click (fn [] nil)}]
         [:button {:title "Play" :class (compose (s/no-select) (s/circle-bounding)
                                                 (s/player-button) (s/player-play-button) "la"
                                                 (if (@state :paused) "la-play" "la-pause"))
                   :on-click (play-pause! state)
                   :on-key-press (fn [e] (if (and (= data/space-char (-> e .-charCode)) (.-stopPropagation e)) (.stopPropagation e)))}]
         [:button {:title "Previous" :class (compose (s/no-select) (s/circle-bounding)
                                                     (s/player-button) "la la-fast-forward")
                   :on-click (fn [] )}]]
        
        ;; info area
        [:div {:class (compose (s/player-info-area))}
         [:div (get-in @state [:song :title])]
         [:div (get-in @state [:album :title])]
         [:div (get-in @state [:artist :name])]]
        
        ;; volume
        [:div {:class (compose (s/right) (s/player-volume-area))}
         [:input {:type "range" :min "0" :max "1" :step "0.001" :value @data/volume
                  :class (compose (s/player-volume-slider 6))
                  :on-change (fn [e] (let [v (-> e .-target .-value)]
                                      (when (-> @state :audio nil? not)
                                        (set! (-> @state :audio .-volume) v))
                                      (reset! data/volume v)))}]]]

       ;; audio element
       [audio state audio-elem player-id]])))


