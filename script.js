async function nextEpisode(id) {
    const url = "/update?id=" + id

    const response = await fetch(url, { method: "POST" })

    location.reload()
}