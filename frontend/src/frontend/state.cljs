(ns frontend.state
  (:require [reagent.core :as r :refer [atom]]))

(def state (r/atom {}))
