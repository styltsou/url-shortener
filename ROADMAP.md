# URL Shortener Feature Roadmap

## Target Market
Solo founders and SaaS builders who need reliable link management with analytics. Focus on developers building in public and indie hackers who value transparency and data-driven decisions.

---

## Phase 1: Foundation & Core Features (v1.0 Release)

### 1. Link Tagging System
**Goal:** Allow users to organize and categorize their links for better management.

**Time Estimate:** 20-25 hours
- Migrations & SQL queries: 6-8 hours (learning SQL, testing edge cases)
- Backend API endpoints: 4-5 hours (SQLC generation, handler logic)
- Frontend UI: 8-10 hours (tag management page, selectors, filters - can vibe code most of this)
- Testing & bug fixes: 2-3 hours

**Database Design:**
- `tags` table: user-owned tags with unique names per user
- `link_tags` junction table: many-to-many relationship
- Soft deletes on tags for data integrity

**Features:**
- Create, read, update, delete tags
- Assign multiple tags to a single link
- Remove tags from links
- View all tags in a central management interface
- Filter links by single tag
- Filter links by multiple tags (AND/OR logic)

**UI Components:**
- Tag management page (CRUD operations)
- Tag selector/input when creating/editing links
- Tag badges on link cards
- Tag filter dropdown/multiselect on links list
- Tag pills with remove functionality on individual links

**API Endpoints:**
```
GET    /api/tags              - List user's tags
POST   /api/tags              - Create new tag
PATCH  /api/tags/:id          - Update tag
DELETE /api/tags/:id          - Delete tag
POST   /api/links/:id/tags    - Add tags to link
DELETE /api/links/:id/tags    - Remove tags from link
GET    /api/links?tags=x,y    - Filter links by tags
```

**SQL Queries to Implement:**
- Fetch links with all their tags (JOIN query)
- Filter links by tag IDs (WHERE IN clause)
- Count links per tag (GROUP BY aggregation)
- Find links with multiple specific tags (HAVING clause)
- Delete orphaned tags (tags with no links)

**Testing Considerations:**
- Tag name uniqueness per user
- Preventing duplicate tag assignments
- Cascade behavior when deleting tags
- Soft delete implications

---

### 2. QR Code Generation (v1)
**Goal:** Generate QR codes for shortened links that users can download and use in marketing materials.

**Time Estimate:** 12-15 hours
- Backend QR generation: 4-5 hours (library integration, file handling)
- Redis caching setup: 2-3 hours (first time using Redis with Go)
- API endpoints: 2 hours
- Frontend UI: 4-5 hours (modal, preview, download - mostly vibe code)

**Core Functionality:**
- Generate QR code on-demand for any shortened link
- Support multiple formats: PNG, SVG
- Support multiple sizes: Small (256x256), Medium (512x512), Large (1024x1024)
- Error correction levels: Low, Medium, High
- Basic black and white styling

**Technical Implementation:**
- Use Go QR code library (e.g., `github.com/skip2/go-qrcode`)
- Generate QR codes server-side
- Cache generated QR codes in Redis with key: `qr:{shortcode}:{format}:{size}`
- Set Redis TTL to 24 hours (regenerate daily or on link update)
- Store QR code generation count in ClickHouse analytics
- Invalidate Redis cache when link destination changes

**Features:**
- Generate button on link detail page
- Preview QR code before download
- Download in selected format and size
- QR code points to shortened URL (not original)
- Regenerate if link destination changes

**UI Components:**
- QR code preview modal
- Format selector (PNG/SVG radio buttons)
- Size selector (dropdown or buttons)
- Download button
- Copy QR code URL button

**API Endpoints:**
```
GET /api/links/:shortcode/qr?format=png&size=512 - Generate/download QR
GET /qr/:shortcode                               - Public QR endpoint
```

**Free Tier Limitations:**
- Plain black/white QR codes only
- Standard error correction
- Basic formats (PNG, SVG)

**Future Premium Features (note for later):**
- Branded QR codes with logo in center
- Custom colors
- Custom shapes/styles
- High-resolution outputs

---

### 3. Basic Analytics
**Goal:** Provide actionable insights about link and QR code performance.

**Time Estimate:** 35-45 hours
- ClickHouse setup & schema: 6-8 hours (learning ClickHouse, migrations)
- Click tracking middleware: 4-5 hours (parsing user agents, IP hashing)
- Simple goroutine implementation: 2 hours
- Batching with channels (v1.1): 4-6 hours (learning Go channels deeply)
- Backend analytics queries: 8-10 hours (learning ClickHouse SQL, aggregations, time windows)
- API endpoints for analytics: 3-4 hours
- Frontend dashboard: 10-12 hours (charts, tables, date pickers - can vibe code with recharts)
- Testing & optimization: 3-4 hours

**Data Collection:**
- Click timestamp
- Referrer URL
- User agent (for device/browser detection)
- IP address (for geographic data, hashed for privacy)
- Click source: direct link vs QR code scan

**Database Design (ClickHouse):**
- `clicks` table (columnar storage optimized for analytics):
  - id (UUID)
  - link_id (UUID)
  - clicked_at (DateTime)
  - referrer (String, nullable)
  - user_agent (String)
  - ip_hash (FixedString(64))
  - country_code (FixedString(2))
  - device_type (Enum8: 'desktop', 'mobile', 'tablet', 'bot')
  - browser (String)
  - source_type (Enum8: 'link', 'qr')
  - created_at (DateTime)

**Why ClickHouse:**
- Extremely fast aggregations and time-series queries
- Excellent compression (10x-100x vs Postgres)
- Built for analytical workloads (GROUP BY, time windows)
- Can handle millions of clicks efficiently

**ClickHouse Query Patterns:**
- Use materialized views for pre-aggregated daily/hourly stats
- Leverage ORDER BY (link_id, clicked_at) for fast filtering
- Use DateTime functions for time bucketing
- Projection queries for device/browser breakdowns

**Analytics to Display:**

**Overview Metrics (per link):**
- Total clicks
- Unique clicks (based on IP hash + user agent combo)
- Click-through rate over time
- QR scans vs direct link clicks

**Time Series:**
- Clicks over time graph (last 7 days, 30 days, 90 days, all time)
- Hourly breakdown for last 24 hours
- Daily breakdown for selected range
- Peak activity times

**Referrer Analysis:**
- Top 10 referrers
- Direct vs referred traffic ratio
- Social media breakdown (Twitter, Facebook, LinkedIn, etc.)
- Unknown/other referrers grouped

**Geographic Data:**
- Top 5 countries by clicks
- World map visualization (optional for v1, can be v1.1)

**Device & Browser:**
- Desktop vs Mobile vs Tablet split (pie chart)
- Top 5 browsers
- Bot traffic identification and filtering

**UI Components:**
- Analytics dashboard for individual links
- Date range selector
- Summary cards (total clicks, unique visitors, QR scans)
- Line chart for click trends
- Bar chart for top referrers
- Pie chart for device types
- Table for detailed click log (paginated, last 100 clicks)

**Performance Considerations:**
- ClickHouse handles aggregations natively, no need for separate aggregation tables
- Use materialized views in ClickHouse for pre-computed metrics
- Cache aggregated results in Redis with TTL (5-15 minutes)
- Redis cache keys: `analytics:{link_id}:{metric}:{timerange}`
- Index on (link_id, clicked_at) in ClickHouse for fast filtering
- **Click Recording Strategy (Progressive Implementation):**
  - **v1.0 - Simple Goroutine:** Use `go recordClick()` for non-blocking inserts, one click at a time
  - **v1.1 - Batching with Channels:** Buffer clicks in memory channel, batch insert every N clicks or M seconds (learn: channels, select, tickers)
  - **v1.2 - Redis Queue:** Replace in-memory channel with Redis queue for durability and distributed processing (learn: message queues, fault tolerance)
- Use ClickHouse's built-in time bucketing functions

**API Endpoints:**
```
POST   /api/clicks                    - Record a click (internal)
GET    /api/links/:id/analytics       - Get analytics summary
GET    /api/links/:id/analytics/time  - Time series data
GET    /api/links/:id/analytics/referrers - Referrer breakdown
GET    /api/links/:id/clicks          - Recent clicks log
```

**Privacy Considerations:**
- Hash IP addresses, don't store raw IPs
- Aggregate data after 90 days, delete individual records
- GDPR-compliant data retention policy
- Allow users to opt-out of tracking (future feature)

---

### 4. Custom Link Preview (Open Graph Tags)
**Goal:** Let users customize how their shortened links appear when shared on social media.

**Time Estimate:** 18-22 hours
- Database migrations: 1 hour
- OG tag scraper: 4-5 hours (HTTP client, parsing HTML)
- Bot detection logic: 2-3 hours (user agent parsing, conditional serving)
- Backend API: 3-4 hours
- Frontend editor UI: 6-8 hours (form, preview cards - can vibe code most of this)
- Testing with actual social media crawlers: 2-3 hours

**Why This Matters:**
- Most URL shorteners just use the destination page's OG tags
- Custom previews let users control messaging
- Essential for marketing campaigns and A/B testing messaging
- Differentiator from competitors

**Database Changes:**
- Add to `links` table:
  - og_title (varchar 255)
  - og_description (text)
  - og_image_url (varchar 2048)
  - use_custom_preview (boolean, default false)

**Features:**
- Toggle between default (destination) and custom preview
- Set custom title (up to 60 chars recommended)
- Set custom description (up to 160 chars recommended)
- Upload or provide URL for custom image
- Preview how it will look on different platforms (Twitter, Facebook, LinkedIn)
- Fallback to destination OG tags if custom not set

**Technical Implementation:**
- Fetch destination URL's OG tags as default (background job)
- Store custom OG tags in database
- Serve custom OG tags from shortcode URL's `<head>`
- Cache OG tag responses for performance
- Validate image URLs and sizes

**Image Handling:**
- Allow URL input for images (external hosting)
- Optional: Upload to your own storage (S3/CDN)
- Validate image format (JPEG, PNG, WebP)
- Recommend 1200x630px for best compatibility
- Show warnings for non-standard sizes

**UI Components:**
- OG tag editor on link create/edit page
- Toggle switch: "Use custom preview"
- Title input with character counter
- Description textarea with character counter
- Image URL input or file upload
- Live preview cards showing Twitter/Facebook/LinkedIn appearance
- "Test preview" button to see how link renders

**API Endpoints:**
```
GET    /api/links/:id/preview         - Get preview data
PATCH  /api/links/:id/preview         - Update custom preview
POST   /api/links/:id/preview/fetch   - Fetch destination OG tags
DELETE /api/links/:id/preview         - Remove custom preview
```

**Server-Side Rendering:**
- For shortcode URLs (e.g., `/abc123`), serve HTML with OG tags before redirect
- Use bot detection to determine if request is from social media crawler
- If bot: serve page with OG tags + meta refresh
- If human: immediate 302 redirect
- Popular crawlers: facebookexternalhit, Twitterbot, LinkedInBot

**Free vs Premium Considerations:**
- Free: Custom title and description only
- Premium: Custom images, multiple preview variants for A/B testing

---

## Phase 2: Polish & Iteration

### Refinements to Existing Features

**Tags Enhancement:**
- Bulk tag operations (add/remove tags to multiple links)
- Tag colors for visual organization
- Tag usage statistics (most used tags)
- Recently used tags quick access
- Tag autocomplete with fuzzy search
- Import/export tags as JSON

**QR Codes Enhancement (Branded QR v1):**
- Add logo/image to center of QR code
- Custom foreground/background colors
- Rounded corners vs square pixels
- Add text label below QR code
- Download as print-ready PDF
- Batch QR generation for multiple links
- QR code templates library

**Analytics Enhancement:**
- Real-time click tracking (WebSocket updates)
- Export analytics data as CSV
- Email reports (daily/weekly/monthly digests)
- Comparison between links (side-by-side analytics)
- Conversion tracking (if destination has tracking pixel)
- Goal setting and alerts (e.g., notify when link hits 1000 clicks)
- Integration with Google Analytics

**Link Preview Enhancement:**
- Multiple preview variants for A/B testing
- Schedule preview changes (change OG tags at specific date/time)
- Preview history and version control
- AI-generated descriptions (using Claude API)
- Bulk update previews for multiple links
- Preview templates library

---

## Phase 3: Advanced Features (Post v1.0)

### Custom Domains
**Complexity:** High
**Value:** Very High for branding

**Features:**
- Users bring their own domain (brand.io)
- DNS verification and setup instructions
- Automatic SSL certificate provisioning (Let's Encrypt)
- Domain-level analytics
- Multiple domains per account (premium tier)
- Subdomain support (links.brand.io)

**Technical Challenges:**
- Dynamic SSL certificate management
- DNS verification flow
- Reverse proxy configuration
- Domain-specific short code namespaces

---

### A/B Testing
**Complexity:** Medium-High
**Value:** High for marketers

**Features:**
- Create multiple destination URLs for one short code
- Define traffic split percentages (50/50, 70/30, etc.)
- Rotation strategies: random, round-robin, weighted
- Track conversion metrics per variant
- Automatic winner detection based on goals
- Schedule A/B test duration

**Use Cases:**
- Test different landing pages
- Test different OG tag previews
- Product launch variants
- Seasonal campaign testing

---

### Team Workspaces
**Complexity:** High
**Value:** High for agencies and growing startups

**Features:**
- Create workspaces/organizations
- Invite team members with roles (admin, editor, viewer)
- Share links within workspace
- Workspace-level analytics
- Audit logs for team actions
- Centralized billing

---

### Link Bundles/Collections
**Complexity:** Low-Medium
**Value:** Medium

**Features:**
- Group multiple links into a collection
- Single URL that displays all links in collection
- Bio link page alternative (like Linktree)
- Customizable landing page for collections
- Track clicks across entire collection

---

### Advanced Security Features
**Features:**
- Password-protected links
- Expiration dates (link expires after date)
- Click limits (deactivate after N clicks)
- Geographic restrictions (only accessible from certain countries)
- IP whitelist/blacklist
- Bot detection and blocking

---

### API & Integrations
**Features:**
- Public REST API with rate limiting
- API keys and OAuth support
- Zapier integration
- Slack bot for link creation
- Chrome/Firefox extension
- iOS/Android sharing extension
- Webhooks for click events

---

### Building in Public Features
**Features:**
- Public stats pages (optional per link)
- Embeddable click counter widgets
- Public analytics dashboards
- Social sharing of milestones
- Leaderboards (opt-in)

---

## Free vs Premium Tier Structure (Suggestion)

### Free Tier:
- 100 links
- Basic analytics (30 days retention)
- Plain QR codes
- Custom link previews (text only)
- 5 tags
- Community support

### Premium Tier ($9-19/month):
- Unlimited links
- Advanced analytics (unlimited retention)
- Branded QR codes with logo
- Custom images in link previews
- Unlimited tags with colors
- Custom domain (1 domain)
- Priority support
- A/B testing (3 variants)
- Export data

### Enterprise Tier ($49+/month):
- Everything in Premium
- Team workspaces (up to 10 members)
- Multiple custom domains
- Advanced security features
- API access
- White-label options
- SLA guarantees

---

## Technical Debt & Infrastructure

### Items to Address During Iteration:
- Implement proper caching strategy (Redis)
- Set up CDN for QR codes and static assets
- Database query optimization and indexing
- Background job queue for heavy operations (QR generation, analytics aggregation)
- Rate limiting on API endpoints
- Comprehensive error handling and logging
- Automated backups and disaster recovery
- Load testing and performance benchmarks
- Security audit and penetration testing

---

## Marketing & Growth Strategy

### Content Ideas:
- Blog about URL shortener best practices
- Case studies from early users
- Building in public documentation
- Technical deep-dives (how you built X feature)
- Comparison with competitors
- SEO optimization for "best URL shortener for [use case]"

### Distribution Channels:
- Product Hunt launch
- Hacker News Show HN
- Indie Hackers community
- Twitter/X building in public thread
- Dev.to and Medium articles
- Reddit (r/SideProject, r/SaaS)

---

## Success Metrics to Track

### Product Metrics:
- Monthly Active Users (MAU)
- Links created per user
- Average clicks per link
- QR code generation rate
- Feature adoption rates
- User retention (Day 7, Day 30)

### Business Metrics:
- Free to paid conversion rate
- Monthly Recurring Revenue (MRR)
- Churn rate
- Customer Acquisition Cost (CAC)
- Lifetime Value (LTV)

### Technical Metrics:
- API response times
- Uptime percentage
- Error rates
- Database query performance
- CDN cache hit rate

---

## Timeline Estimate

**Phase 1 (v1.0 Release):**
- Tagging: 20-25 hours (~3-4 days at 6-8 hours/day)
- QR Codes v1: 12-15 hours (~2 days)
- Basic Analytics: 35-45 hours (~5-6 days)
- Custom Link Previews: 18-22 hours (~3 days)
- Polish & Testing: 15-20 hours (~2-3 days)
**Total: 100-127 hours (~15-18 days of focused work)**

**Phase 2 (Iteration):**
- Branded QR codes: 12-15 hours
- Tag enhancements: 10-12 hours
- Analytics enhancements: 20-25 hours
- Preview enhancements: 8-10 hours
**Total per feature: varies, 8-25 hours each**

**Phase 3 (Advanced Features):**
- Custom domains: 30-40 hours (DNS, SSL, complex)
- A/B testing: 25-30 hours (routing logic, variant tracking)
- Team workspaces: 40-50 hours (auth, permissions, complex)
- Other features: 10-30 hours each depending on complexity

---

## Next Steps

1. ✅ Complete tagging system migrations and queries
2. ✅ Build tag management UI
3. ✅ Implement QR code generation backend
4. Set up analytics database schema
5. Build click tracking middleware
6. Create analytics dashboard UI
7. Implement OG tag scraping and storage
8. Build link preview editor
9. End-to-end testing of all features
10. Deploy v1.0 and gather user feedback

---

## Notes & Considerations

- Focus on shipping v1.0 quickly, iterate based on real user feedback
- Prioritize features that differentiate you from Bitly/Dub.co
- Keep code quality high - this is a portfolio piece
- Document architecture decisions for interviews
- Consider open-sourcing parts of it (build credibility)
- Build with scale in mind but don't over-engineer for Day 1
- Perfect is the enemy of done - ship and improve