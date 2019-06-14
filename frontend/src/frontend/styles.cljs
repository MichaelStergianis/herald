(ns frontend.styles
  (:require [clojure.string :refer [join]]
            [cljss.core :as css :refer-macros [defstyles defkeyframes]]
            [cljss.reagent :refer-macros [defstyled]]))

(defn compose [& styles]
  (join " " styles))

(def primary      "#4527a0")
(def p-light      "#7953d2")
(def p-dark       "#000070")
(def secondary    "#b39ddb")
(def s-light      "#e6ceff")
(def s-dark       "#836fa9")
(def white        "#ffffff")
(def highlighted  "#e8e8e8")
(def black        "#000000")
(def green        "#4CAF50")
(def border-green "#1B5E20")
(def red          "#e53935")
(def border-red   "#b71c1c")

(def bg-primary {:background-color primary :color white})
(def bg-secondary {:background-color secondary :color white})

(defstyles roboto-font []
  {:font-family "Roboto, sans-serif"})

(defstyles pad-in-start [padding]
  {:padding-left (str padding "px")})

(defstyles no-select []
  {:-webkit-touch-callout "none" #_(iOS Safari)
     :-webkit-user-select "none" #_(Safari)
      :-khtml-user-select "none" #_(Konqueror HTML)
        :-moz-user-select "none" #_(Firefox)
         :-ms-user-select "none" #_(Internet Explorer/Edge)
             :user-select "none" #_(Chrome and Opera)})


(defstyles border [style]
  {:border-style style})

(defstyles margin [margin]
  {:margin margin})

(defstyles display [display]
  {:display display})

(defstyles vertical-align [align]
  {:vertical-align align})

(defstyles bg [color border-color]
  {:background-color color
   :border-color border-color})

(defstyles back-toggle []
  {:cursor "pointer"})

(defstyles navbar []
  {:position "fixed"
   :z-index "1000"
   :top "0"
   :left "0"
   :background-color primary
   :color white
   :display "table"
   :width "100%"})

(defstyles above-nav [height]
  {:background-color p-dark
   :height (str height "px")
   :width "100%"})

(defstyles between-above-nav [height]
  {:background-color p-light
   :height (str height "px")
   :width "100%"})

(defstyles navbar-nav [height]
  {:padding "8px"
   :height (str height "px")
   :display "inline"})

(defstyles navbar-toggle [height-and-width]
  {:display "inline"
   :height (str height-and-width "px")
   :width  (str height-and-width "px")
   :text-align "center"
   :background-color "inherit"
   :cursor "pointer"
   :outline "none"
   :border "none"
   :font-size "18px"
   :color white})

(defstyles toggle []
  {:padding "4px"
   :transition "background-color 0.1s ease-in-out"})

(defstyles circle-bounding []
  {:border-radius "50%"
   :-webkit-shape-outside "circle()"
   :shape-outside "circle()"})

(defstyles color-on-active [color]
  {:background-color color})

(defstyles navbar-brand []
  {:height "inherit"
   :font-size "18px"
   :cursor "pointer"})

(defstyles sidebar [document-height top width]
  {:position "fixed"
   :width (str width "px")
   :top (str top "px")
   :left (str "-" (+ width (/ width 10)) "px")
   :height (str (- document-height top) "px")
   :background-color white
   :transition "left 0.5s"
   :overflow-x "hidden"
   :box-shadow "1px 0px 3px 0px #c0c0c0"
   :z-index "1001"})

(defstyles sidebar-open []
  {:display "block"
   :left "0px"})

(defstyles sidebar-ul []
  {:padding-inline-start "0px"
   :font-size "15px"
   :margin-top "0px"
   :margin-bottom "0px"
   :list-style-type "none"})

(defstyles menu-li []
  {:cursor "pointer"
   :padding "12px"})

(defstyles highlighted-row []
  {:background-color highlighted})

(defstyles sidebar-li-icon []
  {:font-size "18px"})

(defstyles sidebar-li-a []
  {})

(defstyles left []
  {:float "left"})

(defstyles right []
  {:float "right"})

(defstyles sr-only []
  {:border "0"
   :clip "rect(1px, 1px, 1px, 1px)"
   :clip-path "inset(50%)"
   :height "1px"
   :margin "-1px"
   :overflow "hidden"
   :padding "0"
   :position "absolute"
   :width "1px"
   :word-wrap "normal !important"})

(defstyles album [wh padding]
  {:width  (str wh "px")
   :height (str wh "px")
   :padding (str padding "px")
   :display "inline-block"
   :position "relative"
   :margin "8px"
   :box-shadow "2px 2px 4px 1px #c0c0c0"
   :visibility "visible"
   :cursor "pointer"
   })

(defstyles album-inside []
  {})

(defstyles album-info []
  {:position "relative"
   :float "left"
   :z-index "3"
   :text-align "left"
   :color white})

(defstyles album-background []
  {:width  "100%"
   :height "100%"
   :position "absolute"
   :z-index "1"
   :top "0"
   :left "0"
   :background-color secondary})

(defstyles album-img []
  {:width  "100%"
   :height "100%"
   :margin "0"
   :top "50%"
   :color s-light
   :transform "translateY(-25%)"
   :-ms-transform "translateY(-25%)"
   :font-size "84px"
   :text-align "center"
   :z-index "2"
   :position "absolute"})

(defstyles album-buttons []
  {:z-index "3"
   :display "inline"
   :position "absolute"
   :bottom "0"
   :right "0"
   :font-size "18px"
   :color white
   :visibility "hidden"
   :opacity "0"
   :transition "visibility 0.3s, opacity .3s"})

(defstyles album-buttons-show []
  {:visibility "visible"
   :opacity "1"})

(defstyles album-button [padding margin]
  {:padding (str padding "px")
   :margin  (str margin "px")
   :background-color s-light
   :transition "color .07s ease-in-out"})

(defstyles album-button-clicked []
  {:color s-dark})

(defstyles full-screen-backdrop [width height z-index]
  {:position "fixed"
   :top "0"
   :left "0"
   :display "none"
   :width  (str width "px")
   :height (str height "px")
   :z-index (str z-index)})

(defstyles full-screen-backdrop-active []
  {:display "block"
   :cursor "default"})

(defstyles options-menu [top right]
  {:display "block"
   :position "absolute"
   :z-index "99"
   :visibility "hidden"
   :opacity "0"
   :background-color white
   :border-radius "2px"
   :color black
   :box-shadow "1px 0px 3px 0px #c0c0c0"
   :transition "visibility 0.1s, opacity 0.1s ease-in-out"
   :&:hover {:background-color highlighted}
   :top "16px"
   :right "8px"})

(defstyles options-menu-active []
  {:visibility "visible"
   :opacity "1"})

(defstyles context-menu-shadow []
  {:box-shadow "2px 2px 4px 2px #c0c0c0a0"})

(defstyles manage-library-menu [width height z-index]
  {:max-height "40%"
   :cursor "default"
   :z-index (str z-index)
   :background-color white
   :border-radius "4px"
   :color black})

(defstyles manage-library-row []
  {:display "grid"
   :grid-template-columns "25% auto 40px 40px"
   :grid-column-gap "4px"
   :height "40px"
   :padding "4px 0"
   :width "100%"})

(defstyles grid-column [column]
  {:grid-column column})

(defstyles button []
  {:cursor "pointer"
   :outline "none"
   :border-style "solid"
   :border-width "2px"
   :border-radius "4px"
   :padding "6px"
   :color black
   :background-color white})

(defstyles manage-lib-cell []
  {:padding "8px"
   :border-width "1px"
   :border-radius "4px"
   :font-size "16px"
   :cursor "auto"
   :overflow "auto"
   :text-align "left"})

(defstyles hr []
  {:margin "0"})

(defstyles settings []
  {:text-align "center"
   :font-size "16px"})

(defstyles setting []
  {:padding "16px 0"
   :cursor "pointer"
   :&:hover {:background-color highlighted}})
