(defproject frontend "0.1.0-SNAPSHOT"
  :description "Music server frontend"
  :url "http://music.stergianis.ca"
  :license {:name "GPL V2.0"
            :url "https://www.gnu.org/licenses/old-licenses/gpl-2.0.en.html"}
  
  :min-lein-version "2.7.1"

  :dependencies [[org.clojure/clojure "1.9.0"]
                 [org.clojure/clojurescript "1.10.312"]
                 [org.clojure/core.async  "0.4.474"]
                 [reagent "0.8.1"]
                 [cljs-ajax "0.7.5"]]

  :source-paths ["src"]

  :resource-paths ["resources" "target"]

  :cljsbuild {:builds
              [{:id "dev"
                :source-paths ["src"]

                ;; The presence of a :figwheel configuration here
                ;; will cause figwheel to inject the figwheel client
                ;; into your build
                :figwheel-main {:on-jsload "frontend.core/on-js-reload"
                           ;; :open-urls will pop open your application
                           ;; in the default browser once Figwheel has
                           ;; started and compiled your application.
                           ;; Comment this out once it no longer serves you.
                           :open-urls ["http://localhost:8080"]}

                :compiler {:main frontend.core
                           :asset-path "js/compiled/out"
                           :output-to "resources/public/js/compiled/frontend.js"
                           :output-dir "resources/public/js/compiled/out"
                           :source-map-timestamp true
                           ;; To console.log CLJS data-structures make sure you enable devtools in Chrome
                           ;; https://github.com/binaryage/cljs-devtools
                           :preloads [devtools.preload]}}
               ;; This next build is a compressed minified build for
               ;; production. You can build this with:
               ;; lein cljsbuild once min
               {:id "min"
                :source-paths ["src"]
                :compiler {:output-to "resources/public/js/compiled/frontend.js"
                           :main frontend.core
                           :optimizations :none
                           :pretty-print false}}]}

  ;; Setting up nREPL for Figwheel and ClojureScript dev
  ;; Please see:
  ;; https://github.com/bhauman/lein-figwheel/wiki/Using-the-Figwheel-REPL-within-NRepl
  :profiles {:dev {:dependencies [[com.bhauman/figwheel-main "0.2.0"]
                                  [com.bhauman/rebel-readline-cljs "0.1.4"]
                                  [cider/piggieback "0.4.0"]
                                  [org.clojure/clojurescript "1.10.339"]]
                   :source-paths ["src"]

                   ;; for CIDER
                   ;; :plugins [[cider/cider-nrepl "0.21.1"]]
                   ;; :repl-options {:nrepl-middleware [[cider.piggieback/wrap-cljs-repl]]}
                   ;; need to add the compliled assets to the :clean-targets
                   :clean-targets ^{:protect false} ["resources/public/js/compiled"
                                                     :target-path]}})
