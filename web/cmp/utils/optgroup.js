/**
 * OptGroup Mithril component.
 * Renders <optgroup> and <option> elements for a <select>.
 * 
 * @param {Array} rows - Array of data objects.
 * @param {string} groupCol - Property name for grouping (optgroup label).
 * @param {string} contentCol - Property name for option display text.
 * @param {string} valueCol - Property name for option value.
 */
const OptGroup = {
    view: ({ attrs: { rows, groupColname, contentColname, valueColname } }) => {
        if (!Array.isArray(rows) || !groupColname || !contentColname || !valueColname) return null;

        // Group rows by groupColname
        const groups = rows.reduce((acc, row) => {
            const group = row[groupColname] || "";
            if (!acc[group]) acc[group] = [];
            acc[group].push(row);
            return acc;
        }, {});

        return Object.entries(groups).map(([group, items]) =>
            group
                ? m("optgroup", { label: group },
                    items.map(item =>
                        m("option", { value: item[valueColname] }, item[contentColname])
                    )
                )
                : items.map(item =>
                    m("option", { value: item[valueColname] }, item[contentColname])
                )
        );
    }
};