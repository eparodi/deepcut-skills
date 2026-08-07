# Live Streaming Platform

**Status:** Review
**Owner:** Eliseo
**Created:** 2026-08-06

## Requirements

### User Story 1: Streamer Onboarding (US1)

As a streamer, I want to sign up with my Google account and get a stream
key so that I can start broadcasting from OBS without friction.

**Acceptance Criteria:**

- Given I am on the landing page, When I click "Start Streaming", Then I see a "Sign in with Google" button
- Given I click "Sign in with Google", When I authenticate with my Google account, Then I am redirected to my dashboard
- Given I am a newly authenticated streamer, When I land on my dashboard, Then I see my unique stream key prominently displayed with a copy button
- Given I am on my dashboard, When I click "Regenerate Stream Key", Then a confirmation dialog appears, and on confirm my old key is revoked and a new one is generated
- Given I have my stream key, When I configure OBS with `rtmp://<server>/live` and my key, Then OBS connects and I see a green connection indicator

### User Story 2: Going Live (US2)

As a streamer, I want to start and stop my stream so that viewers know
when I'm live and can watch in real time.

**Acceptance Criteria:**

- Given OBS is streaming to the RTMP ingest, When the server receives the stream, Then my channel status changes to "Live" and appears on the homepage live list
- Given I am live, When I stop streaming in OBS, Then my channel status changes to "Offline" within 30 seconds and the stream ends gracefully
- Given I am live, When a viewer navigates to my channel page, Then they see the live video player with HLS playback
- Given I am live, When the stream drops unexpectedly (OBS crash, network loss), Then my channel shows "Stream Interrupted" for up to 60 seconds before showing "Offline"

### User Story 3: Viewing Live Streams (US3)

As a viewer, I want to browse live streams and watch one so that I can
discover content in real time.

**Acceptance Criteria:**

- Given I am on the homepage, When there are active streams, Then I see a grid of live channels with thumbnail, streamer name, title, and viewer count
- Given I am on the homepage, When I click on a live channel, Then I navigate to `/channel/<username>` where the HLS player loads and starts playing within 3 seconds
- Given I am watching a stream, When the streamer is active, Then the video plays with no more than 5 seconds of latency from the live broadcast
- Given there are no live streams, When I visit the homepage, Then I see an empty state: "No one is live right now. Check out past streams below."

### User Story 4: Real-Time Chat (US4)

As a viewer, I want to send and read chat messages alongside the stream
so that I can participate in the conversation.

**Acceptance Criteria:**

- Given I am watching a live stream, When the streamer is live, Then I see a chat panel alongside the video player showing real-time messages
- Given I am signed in with Google, When I type a message and press Enter, Then my message appears in the chat within 1 second
- Given I am NOT signed in, When I try to send a chat message, Then I see "Sign in to chat" linking to Google OAuth
- Given I am watching a past stream (VOD), When I view the video page, Then the chat replay scrolls alongside the video timeline

### User Story 5: Past Streams / VOD (US5)

As a viewer, I want to search and watch past streams so that I can catch
up on content I missed or discover new streamers.

**Acceptance Criteria:**

- Given a stream has ended, When I visit the streamer's channel page, Then I see a list of their past streams with title, date, duration, and thumbnail
- Given I am on the homepage or search page, When I search by keyword, Then I see past streams matching the title or streamer name
- Given I click on a past stream, When the VOD page loads, Then the HLS recording plays back with standard video controls (play, pause, seek, volume)
- Given a stream has ended, When the recording is being processed, Then the VOD shows "Processing — available soon" for up to 5 minutes before becoming playable

### User Story 6: Streamer Dashboard (US6)

As a streamer, I want to manage my stream settings and see basic
analytics so that I can control how my stream appears.

**Acceptance Criteria:**

- Given I am on my dashboard, When I set a stream title and category, Then it updates and appears on my channel page immediately
- Given I am on my dashboard, When I view my analytics panel, Then I see: total stream time (this week), peak concurrent viewers, total unique viewers
- Given my stream is live, When I click "End Stream" on my dashboard, Then the RTMP connection is terminated and the stream ends

## Non-Goals

- ❌ **Monetization** — subscriptions, bits/cheers, ads, tipping (v2)
- ❌ **Native mobile apps** — iOS/Android apps (v2, web-only for v1)
- ❌ **Email/password auth** — Google OAuth only for v1
- ❌ **Streamer categories/tags beyond a single text field** — no curated taxonomy yet
- ❌ **Moderation tools** — banning users, deleting messages (v2)
- ❌ **Following/subscribing to channels** — no notification system yet
- ❌ **Clips/highlights** — no clip creation from VODs
- ❌ **Multiple stream keys per user** — one key per account
- ❌ **Transcoding/adaptive bitrate** — single quality output for v1
- ❌ **Embedded player for external sites** — watch on our platform only
- ❌ **CDN** — single origin server for v1 (add CDN when viewer demand requires it)
- ❌ **Streamer profile customization** — bio, avatar, banner are saved for v2 (Google avatar is auto-imported)

## Open Questions

- ✅ RTMP server: SRS (Simple Realtime Server) — modern, lower latency, better HLS support than nginx-rtmp
- ✅ VOD retention: forever (no auto-delete)
- ✅ Viewer count: shown publicly, includes anonymous viewers
- ✅ Landing page: live-stream grid is the homepage (not a marketing page)
