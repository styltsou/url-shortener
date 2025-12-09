CREATE TABLE link_tags (
	link_id UUID NOT NULL,
	tag_id UUID NOT NULL,

	PRIMARY KEY (link_id, tag_id),
	FOREIGN KEY (link_id) REFERENCES links(id) ON DELETE CASCADE,
	FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Index for "get all tags for a link"
CREATE INDEX idx_link_tags_link_id ON link_tags(link_id);

-- Index for "get all links with a certain tag"
CREATE INDEX idx_link_tags_tag_id ON link_tags(tag_id);
