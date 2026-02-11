const responseNode = document.getElementById("response");
const wasmStatus = document.getElementById("wasmStatus");

function pretty(data) {
  return JSON.stringify(data, null, 2);
}

function setResponse(data) {
  responseNode.textContent = pretty(data);
}

function readFormInputs() {
  return {
    idInstance: document.getElementById("idInstance").value,
    apiTokenInstance: document.getElementById("apiTokenInstance").value,
    sendMessageChatId: document.getElementById("sendMessageChatId").value,
    sendMessageText: document.getElementById("sendMessageText").value,
    sendFileChatId: document.getElementById("sendFileChatId").value,
    sendFileUrl: document.getElementById("sendFileUrl").value,
  };
}

function buildRequestViaWasm(method) {
  if (typeof window.goBuildRequest !== "function") {
    throw new Error("WASM helper goBuildRequest is not available");
  }

  const form = readFormInputs();
  const request = window.goBuildRequest(
    form.idInstance,
    form.apiTokenInstance,
    method,
    form.sendMessageChatId,
    form.sendMessageText,
    form.sendFileChatId,
    form.sendFileUrl,
  );

  if (!request || typeof request !== "object") {
    throw new Error("WASM вернул некорректный формат запроса");
  }

  if (request.error) {
    throw new Error(String(request.error));
  }

  return request;
}

async function callMethod(method) {
  const request = buildRequestViaWasm(method);

  setResponse({ loading: true, request });

  const res = await fetch("/api/call", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(request),
  });

  const data = await res.json();
  if (!res.ok || data.error) {
    throw new Error(data.error || `HTTP ${res.status}`);
  }

  setResponse(data.result);
}

function bindButtons() {
  document.querySelectorAll("button[data-method]").forEach((button) => {
    button.addEventListener("click", async () => {
      const method = button.dataset.method;
      try {
        button.disabled = true;
        await callMethod(method);
      } catch (error) {
        setResponse({
          error: error instanceof Error ? error.message : String(error),
        });
      } finally {
        button.disabled = false;
      }
    });
  });
}

function setButtonsDisabled(disabled) {
  document.querySelectorAll("button[data-method]").forEach((button) => {
    button.disabled = disabled;
  });
}

async function initWasm() {
  if (typeof Go === "undefined") {
    throw new Error("wasm_exec.js is not loaded");
  }

  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(fetch("/assets/main.wasm"), go.importObject);
  go.run(result.instance);
  wasmStatus.textContent = "WASM: on (Go required)";
}

bindButtons();
setButtonsDisabled(true);
setResponse({ message: "ожидание запроса" });

initWasm()
  .then(() => setButtonsDisabled(false))
  .catch((error) => {
    wasmStatus.textContent = "WASM: failed";
    setResponse({
      error: error instanceof Error ? error.message : String(error),
      hint: "Соберите WASM: make wasm",
    });
  });
