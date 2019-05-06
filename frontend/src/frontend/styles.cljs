(ns frontend.styles
  (:require [clojure.string :refer [join]]
            [cljss.core :as css :refer-macros [defstyles defkeyframes]]
            [cljss.reagent :refer-macros [defstyled]]))

(defn compose [& styles]
  (join " " styles))

(def primary   "#4527a0")
(def p-light   "#7953d2")
(def p-dark    "#000070")
(def secondary "#b39ddb")
(def s-light   "#e6ceff")
(def s-dark    "#836fa9")
(def white     "#ffffff")
(def black     "#000000")

(def bg-primary {:background-color primary :color white})
(def bg-secondary {:background-color secondary :color white})

(defstyles pad-in-start [padding]
  {:padding-left (str padding "px")})

(defstyles navbar []
  {:position "absolute"
   :top "0px"
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
  {:height "48px"
   :width  "48px"
   :background-color "inherit"
   :cursor "pointer"
   :outline "none"
   :border "none"
   :font-size "18px"
   :color white})

(defstyles circle-bounding []
  {:border-radius "50%"
   :padding "4px"
   :-webkit-shape-outside "circle()"
   :shape-outside "circle()"
   :transition "background-color 0.1s ease-in-out"})

(defstyles color-on-active [color]
  {:background-color color})

(defstyles navbar-brand []
  {:height "inherit"
   :font-size "18px"
   :cursor "pointer"})

(defstyles sidebar [document-height top width]
  {:position "absolute"
   :width (str width "px")
   :top (str top "px")
   :left (str "-" width "px")
   :height (str (- document-height top) "px")
   :background-color white
   :transition "left 0.5s"
   :overflow-x "hidden"
   :box-shadow "1px 0px 3px 0px #c0c0c0"
   :z-index "999"})

(defstyles sidebar-open []
  {:display "block"
   :left "0px"})

(defstyles sidebar-ul []
  {:padding-inline-start "0px"
   :font-size "15px"
   :margin-top "0px"
   :margin-bottom "0px"
   :list-style-type "none"})

(defstyles sidebar-li [padding]
  {:cursor "pointer"
   :padding "12px"})

(defstyles sidebar-li-active []
  {:background-color "#e8e8e8"})

(defstyles sidebar-li-icon []
  {:font-size "18px"})

(defstyles sidebar-li-a []
  {})

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
