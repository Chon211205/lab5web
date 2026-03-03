async function nextEpisode(id) {
    const url = "/update?id=" + id

    const response = await fetch(url, { method: "POST" })

    location.reload()
}

async function deleteSerie(id) {
  const url = "/series?id=" + id;
  await fetch(url, { method: "DELETE" });
  location.reload();
}