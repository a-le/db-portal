/*
// Copyright (C) 2024 https://github.com/a-le
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
const App = {
    conn: "",
    schema: "",
    theme: "",
    tabState: new TabState("result"),
    oninit: () => {
        var theme = localStorage.getItem("theme");
        App.theme = theme ? theme : window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark-mode" : "light-mode";
    },
    view: () => {
        return m("div.grid-main", {
            class: App.theme
        },
            m("header.area-main-header",
                m("nav.brand",
                    m("a", { href: "/", title: "refresh page" }, versionInfo.appName),
                    m("span.ml-10", "v", versionInfo.server),
                )
            ),
            m("section.area-main-menu",
                m("div.grid", { style: "grid-template-columns: auto 1fr auto;" },
                    m("div.grid-col",
                        m("div", m(ConnForm)),   // ConnForm component
                        m("div.ml-10", m(SchemaForm))  // SchemaForm component
                    ),
                    m("div.grid-col.align-items-end.ml-10", m(ConnInfos)), // ConnInfos component
                    m("div.grid-col.ml-auto.mr-0",
                        m("div.ml-10", m(ThemeSwitch)), // ThemeSwitch component
                        m("div.ml-10", m(LogOut)) // LogOut component
                    ),
                ),
            ),
            m("section.area-main-content",
                !App.conn ? null :
                    [
                        m("div.grid-query", {
                            oncreate: function (vnode) {
                                var h0 = document.querySelector('.grid-query').offsetHeight,
                                    h1 = 195, // ideal area-query-editor height, can be greater
                                    h2 = 335, // ideal area-query-output for 15 lines of results
                                    h1 = Math.max(h0 - h2, h1);
                                const LayoutGrid = GridResize('.grid-query', '.area-query-splitter', '.area-query-editor', `${h1}px 3px auto auto`, 195, false);
                                LayoutGrid.init();
                            }
                        },
                            m("section.area-query-editor",
                                m("div.grid-q-editor-datadict", {
                                    oncreate: function (vnode) {
                                        const LayoutGrid = GridResize('.grid-q-editor-datadict', '.area-q-splitter', '.area-q-editor', `1fr 3px 1fr`, 540, true);
                                        LayoutGrid.init();
                                    }
                                },
                                    m("section.area-q-editor", m(QryForm)), // QryForm component
                                    m("div.area-q-splitter.splitter.splitter-vertical"),
                                    m("section.area-q-datadict", m(DataDict)), // DataDict component
                                )
                            ),
                            m("div.area-query-splitter.splitter.splitter-horizontal"),
                            m("div",
                                m("section.area-query-output-menu",
                                    m("div.grid", { style: "grid-template-columns: auto auto 1fr;" },
                                        m("div.grid-col.tab.tab-b", {
                                            class: App.tabState.selectedClass("result"),
                                            onclick: () => App.tabState.set("result")
                                        }, "result"),
                                        m("div.grid-col.tab.tab-b.ml-20", {
                                            class: App.tabState.selectedClass("explain"),
                                            onclick: () => App.tabState.set("explain")
                                        }, "explain"),
                                        m("div.grid-col.align-items-end.ml-50", m(QryInfos)), // QryInfos component
                                    ),
                                ),
                                m("section.area-query-output",
                                    m("div", { class: App.tabState.displayClass("result") }, m(QryResult)), // QryResult component
                                    m("div", { class: App.tabState.displayClass("explain") }, m(QryExplain)), // QryExplain component
                                )
                            )
                        )
                    ]
            ),
            m("footer.area-footer-content.ml-auto.mr-0.text-x-small",
                m("div",
                    m("a", {
                        href: "https://github.com/a-le/" + versionInfo.appName,
                        target: "_blank",
                        title: "View github project page"
                    }, versionInfo.appName
                    ),
                    m("span", m.trust(" &copy; 2024. Licensed under "),
                        m("a", {
                            href: "https://www.gnu.org/licenses/agpl-3.0.html",
                            title: "View License",
                            target: "_blank"
                        }, "AGPL"),
                    ),
                )
            )
        );
    }
};