const COMMENT_CONTENT_MAX_LENGTH = 5000;
const COMMENT_AUTHOR_NAME_MAX_LENGTH = 100;

async function fetchComments(slug) {
  const res = await fetchOrThrow(`${API_BASE}/comments/${encodeURIComponent(slug)}`, {
    label: `comments for slug "${slug}"`,
    notFoundMessage: `No comments available — blog not found for slug "${slug}"`,
  });
  return res.json();
}

async function postComment(slug, payload) {
  const res = await fetch(`${API_BASE}/comments/${encodeURIComponent(slug)}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    const text = (await res.text().catch(() => "")).trim();
    throw new Error(text || `Failed to post comment: ${res.status} ${res.statusText}`);
  }

  return res.json();
}

function formatCommentDate(iso) {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(date);
}

function initComments(container, slug) {
  container.innerHTML = "";

  const list = el("div", "comments__list");
  const form = buildCommentForm({ slug, parentId: null, onSuccess: () => loadAndRenderComments(list, slug) });

  const notice = el(
    "p",
    "comments__notice",
    "Keep it respectful — no hate speech, harassment, or offensive language."
  );

  container.append(el("h2", "comments__heading", "Comments"), list, form, notice);
  loadAndRenderComments(list, slug);
}

async function loadAndRenderComments(listEl, slug) {
  listEl.innerHTML = "";
  listEl.appendChild(el("p", "comments__status", "Loading comments…"));

  try {
    const comments = await fetchComments(slug);
    renderCommentsList(listEl, comments, slug);
  } catch (err) {
    listEl.innerHTML = "";
    listEl.appendChild(el("p", "error-message", err.message || "Failed to load comments."));
  }
}

function renderCommentsList(listEl, comments, slug) {
  listEl.innerHTML = "";

  if (!comments || comments.length === 0) {
    listEl.appendChild(el("p", "comments__empty", "No comments yet — be the first to comment."));
    return;
  }

  comments.forEach((comment) => listEl.appendChild(renderCommentNode(comment, slug, listEl, false)));
}

function renderCommentNode(comment, slug, listEl, isReply) {
  const node = el("div", isReply ? "comment comment--reply" : "comment");

  const header = el("div", "comment__header");
  header.append(
    el("span", "comment__author", comment.authorName),
    el("span", "comment__date", formatCommentDate(comment.createdAt))
  );
  node.append(header, el("p", "comment__content", comment.content));

  if (isReply) return node;

  node.appendChild(buildReplyToggle(comment, slug, listEl));

  if (comment.replies && comment.replies.length > 0) {
    const replies = el("div", "comment__replies");
    comment.replies.forEach((reply) => replies.appendChild(renderCommentNode(reply, slug, listEl, true)));
    node.appendChild(replies);
  }

  return node;
}

function buildReplyToggle(comment, slug, listEl) {
  const toggle = el("button", "comment__reply-toggle", "Reply");
  toggle.type = "button";

  let replyForm = null;
  const close = () => {
    replyForm?.remove();
    replyForm = null;
    toggle.textContent = "Reply";
  };

  toggle.addEventListener("click", () => {
    if (replyForm) {
      close();
      return;
    }
    replyForm = buildCommentForm({
      slug,
      parentId: comment.id,
      showCancel: true,
      onCancel: close,
      onSuccess: () => loadAndRenderComments(listEl, slug),
    });
    replyForm.classList.add("comment-form--reply");
    toggle.insertAdjacentElement("afterend", replyForm);
    toggle.textContent = "Cancel";
  });

  return toggle;
}

function field(labelText, input) {
  const label = el("label", "comment-form__label", labelText);
  label.htmlFor = input.id;
  const wrapper = el("div", "comment-form__field");
  wrapper.append(label, input);
  return wrapper;
}

let commentFormSeq = 0;

function buildCommentForm({ slug, parentId, onSuccess, showCancel = false, onCancel }) {
  const formId = `comment-form-${++commentFormSeq}`;

  const nameInput = el("input", "comment-form__input");
  nameInput.type = "text";
  nameInput.id = `${formId}-name`;
  nameInput.maxLength = COMMENT_AUTHOR_NAME_MAX_LENGTH;

  const anonInput = el("input");
  anonInput.type = "checkbox";
  anonInput.id = `${formId}-anonymous`;
  anonInput.addEventListener("change", () => {
    nameInput.disabled = anonInput.checked;
  });
  const anonLabel = el("label", null, "Post anonymously");
  anonLabel.htmlFor = anonInput.id;
  const anonRow = el("div", "comment-form__checkbox-row");
  anonRow.append(anonInput, anonLabel);

  const contentInput = el("textarea", "comment-form__textarea");
  contentInput.id = `${formId}-content`;
  contentInput.maxLength = COMMENT_CONTENT_MAX_LENGTH;

  const errorMessage = el("p", "comment-form__error error-message");
  errorMessage.hidden = true;

  const submitButton = el("button", "comment-form__submit", parentId ? "Post reply" : "Post comment");
  submitButton.type = "submit";

  const actions = el("div", "comment-form__actions");
  actions.appendChild(submitButton);
  if (showCancel) {
    const cancelButton = el("button", "comment-form__cancel", "Cancel");
    cancelButton.type = "button";
    cancelButton.addEventListener("click", () => onCancel?.());
    actions.appendChild(cancelButton);
  }

  const form = el("form", "comment-form");
  form.append(field("Name", nameInput), anonRow, field("Comment", contentInput), errorMessage, actions);

  const showError = (message) => {
    errorMessage.textContent = message;
    errorMessage.hidden = false;
  };

  let submitting = false;

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (submitting) return;

    const content = contentInput.value.trim();
    const isAnonymous = anonInput.checked;
    const name = nameInput.value.trim();

    if (!content) return showError("Comment cannot be empty.");
    if (content.length > COMMENT_CONTENT_MAX_LENGTH) {
      return showError(`Comment must be ${COMMENT_CONTENT_MAX_LENGTH} characters or fewer.`);
    }
    if (!isAnonymous && !name) return showError('Name is required, or check "Post anonymously".');
    if (!isAnonymous && name.length > COMMENT_AUTHOR_NAME_MAX_LENGTH) {
      return showError(`Name must be ${COMMENT_AUTHOR_NAME_MAX_LENGTH} characters or fewer.`);
    }

    const payload = { content, isAnonymous };
    if (parentId != null) payload.parentId = parentId;
    if (!isAnonymous) payload.authorName = name;

    submitting = true;
    submitButton.disabled = true;
    const originalLabel = submitButton.textContent;
    submitButton.textContent = "Posting…";
    errorMessage.hidden = true;

    try {
      await postComment(slug, payload);
      nameInput.value = "";
      nameInput.disabled = false;
      contentInput.value = "";
      anonInput.checked = false;
      onSuccess?.();
    } catch (err) {
      showError(err.message || "Failed to post comment.");
    } finally {
      submitting = false;
      submitButton.disabled = false;
      submitButton.textContent = originalLabel;
    }
  });

  return form;
}
