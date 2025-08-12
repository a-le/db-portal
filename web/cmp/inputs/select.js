function SelectInput() {
    return {
        view: ({ attrs }) => {
            // Group options by opt.group (undefined group goes to default)
            const grouped = {};
            (attrs.options || []).forEach(opt => {
                const group = opt.group || "";
                if (!grouped[group]) grouped[group] = [];
                grouped[group].push(opt);
            });

            return m("select", {
                name: attrs.name,
                required: attrs.required,
                onclick: attrs.onclick,
                onchange: attrs.onchange,
                onfocus: attrs.onfocus,
                value: attrs.value
            },
                Object.entries(grouped).map(([group, opts]) => {
                    if (group === "") {
                        // No group: render options directly
                        return opts.map(opt =>
                            m("option", { value: opt.value }, opt.label)
                        );
                    } else {
                        // Grouped options: render optgroup
                        return m("optgroup", { label: group },
                            opts.map(opt =>
                                m("option", { value: opt.value }, opt.label)
                            )
                        );
                    }
                })
            );
        }
    };
}

// function SelectInput() {
//     return {
//         view: ({ attrs }) => 
//             m("select", {
//                 onclick: attrs.onclick,
//                 onchange: attrs.onchange,
//                 value: attrs.selected
//             }, 
//                 (attrs.options || []).map(opt =>
//                     m("option", { value: opt.value }, opt.label)
//                 )
//             )
//     };
// }