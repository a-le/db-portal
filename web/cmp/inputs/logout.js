function LogoutInput() {
    return {
        view: () => {
            return [
                m("button[type=button].mb-5", {
                    title: "logout " + App.getUsername(),
                    onclick: function () {
                        App.logout();
                    }
                }, "logout"),
            ]
        }
    }
}
