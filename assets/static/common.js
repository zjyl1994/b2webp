const $each = (selector, callback) => document.querySelectorAll(selector).forEach(callback);
const $eval = expr => (new Function(`return (${expr})`)());
const $onclick = (selector, handler) => $each(selector, e => e.addEventListener('click', handler));

function timestamp2str(ts) {
    let pad2 = num => num < 10 ? '0' + num : num;
    let date = new Date(Number(ts) * 1000);
    let year = date.getFullYear();
    let month = pad2(date.getMonth() + 1);
    let day = pad2(date.getDate());
    let hour = pad2(date.getHours());
    let minute = pad2(date.getMinutes());
    let second = pad2(date.getSeconds());
    return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
}

function formatBytes(bytes) {
    var units = ['B', 'KB', 'MB', 'GB', 'TB'], i;
    for (i = 0; bytes >= 1024 && i < 4; i++) bytes /= 1024;
    return bytes.toFixed(2) + units[i];
}

window.onload = function () {
    $each('[data-render-text]', e => e.innerText = $eval(e.dataset.renderText));
    $each('#main', e => {
        if (e.dataset.title) document.title = e.dataset.title;
        if (e.dataset.navi) $each(`#navitem-${e.dataset.navi}`, e => e.classList.add('active'));
    });
    $onclick('.copy-text', function() {
        try {
            navigator.clipboard.writeText(this.innerText);
        } catch (error) {
            console.error(error.message);
        }
    });
}
