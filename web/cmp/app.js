const JWT_KEY = "jwt";

const App = {
    theme: "",
    dataTransferAction: false,
    claims: {},
    isLogged: () => {
        const jwt = localStorage.getItem(JWT_KEY);
        if (jwt === null)
            return false

        App.claims = parseJwt(jwt);
        if (App.claims.exp && Date.now() / 1000 > App.claims.exp) {
            localStorage.removeItem(JWT_KEY);
            return false
        }
        return true;
    },
    logout: () => {
        localStorage.removeItem(JWT_KEY)
    },
    login: (token) => {
        localStorage.setItem(JWT_KEY, token)
    },
    getAuthHeaders: () => {
        return { "Authorization": "Bearer " + localStorage.getItem(JWT_KEY) }
    },
    getUsername: () => {
        return App.claims.name
    },
    getIsAdmin: () => {
        return App.claims.isadmin === 1 || App.claims.isadmin === "1";
    },
    oninit: () => {
        var theme = localStorage.getItem("theme");
        App.theme = theme ? theme : window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark-mode" : "light-mode";

        App.pageState = new UIState({
            def: window.location.hash.substring(1) || localStorage.getItem("pageState") || "datasources",
            onSet: (tab) => { localStorage.setItem("pageState", tab) }
        });
    },
    view: () => {
        return [
            m("div.grid-main", { class: App.theme },
                m("header.area-main-header.sticky",
                    m("div.grid", { style: "grid-template-columns: auto 1fr auto;" },
                        m("div.grid-col.text-smaller",
                            m("a.no-underline", { href: "/", title: "refresh page" }, versionInfo.AppName),
                            m("span.ml-10", "v", versionInfo.AppVersion),
                        ),
                        m("div.grid-col.ml-auto.mr-auto",
                            App.getIsAdmin() &&
                            [
                                m("a.tab.tab-b.mr-30.no-underline", {
                                    href: "#datasources",
                                    class: App.pageState.selectedClass("datasources"),
                                    onclick: () => {
                                        App.pageState.set("datasources");
                                    }
                                }, m.trust("&#128279;data sources"))
                            ],
                            m("a.tab.tab-b.mr-30.no-underline", {
                                href: "#query",
                                class: App.pageState.selectedClass("query"),
                                onclick: () => {
                                    App.pageState.set("query");
                                }
                            }, m.trust("&#128462;&ThinSpace;SQL editor")),
                            m("a.tab.tab-b.no-underline", {
                                href: "#copy",
                                class: App.pageState.selectedClass("copy"),
                                onclick: () => {
                                    App.pageState.set("copy");
                                }
                            }, m.trust("&#8644;&ThinSpace;copy data")),
                        ),
                        m("div.grid-col.ml-auto.mr-0",
                            m("div.ml-10", m(ThemeSwitch)),
                            m("div.ml-10", m(LogoutInput))
                        ),
                    ),
                ),

                App.isLogged()
                    ? [
                        m("div", { class: App.pageState.displayClass("datasources") }, m(DatasourcesPage)),
                        m("div", { class: App.pageState.displayClass("query") }, m(QueryPage)),
                        m("div", { class: App.pageState.displayClass("copy") }, m(CopyDataPage)),
                    ]
                    : m("div", m(LoginForm)),

                m("footer.area-footer-content.ml-auto.mr-0.text-x-small",
                    m("div",
                        m("a", {
                            href: "https://github.com/a-le/" + versionInfo.AppName,
                            target: "_blank",
                            title: "View github project page"
                        }, versionInfo.AppName
                        ),
                        m("span", m.trust(" &copy; 2025. Licensed under "),
                            m("a", {
                                href: "https://www.gnu.org/licenses/agpl-3.0.html",
                                title: "View License",
                                target: "_blank"
                            }, "AGPL"),
                        ),
                    )
                )
            )
        ];
    }
};
