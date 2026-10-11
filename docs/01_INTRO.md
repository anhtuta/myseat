# My Seat

My Seat is a ticketing system, similar to Ticketmaster. Fans browse events, pick seats on a venue map, hold them through checkout, and receive a ticket. Organizers publish events, price inventory, and open sales.

The hard part is the onsale: many people try to buy the same seats at the same moment, and a seat can be sold only once.

## Tech stack

- Backend: Go
- Frontend: React
- Database: PostgreSQL
- Lock & Cache: Redis
- Virtual Waiting Room: Kafka/RabbitMQ

## Difficult problems

- **Double booking / race conditions.** Two checkout requests can read the same seat as free and both try to sell it. The sale path needs a single winner.
- **Lock contention and timeouts.** Popular sections become hot keys. Long locks stall the queue; short locks let a second buyer slip in after the first request times out.
- **The held-ticket problem.** A seat must leave the map while someone pays, then return if they abandon checkout or the payment fails. Holds expire, and expiry must not race with a successful payment.
- **Fair queueing (the virtual waiting room).** When demand exceeds what checkout can serve, arrivals need an ordered line. Refreshing, opening extra tabs, or reconnecting should not jump the queue.
- **Read path vs. source of truth.** The seat map is read far more often than it is written. A cache makes the map fast, and it must not show a seat as free after it has been held or sold.
- **Payment and inventory disagree.** Charging a card and marking a seat sold are separate steps. A charge without a seat, or a sold seat without a charge, has to be recovered.

## Functional requirements

Fans:

- Browse and search upcoming events by name, date, city, and venue.
- Open an event and see a seating map with section, row, and seat, plus which seats are free, held, or sold.
- Select one or more seats and place a time-limited hold (checkout timer).
- See the price breakdown (ticket price, fees, total) before paying.
- Pay and receive an order confirmation with a scannable ticket.
- View upcoming tickets and past orders on their account.
- Release a hold by leaving checkout, or lose it automatically when the timer ends.

Organizers:

- Create venues with a fixed seat map.
- Create events on a venue, set on-sale and off-sale times, and set prices per section or seat.
- Open and close sales, and see remaining inventory and completed orders.

Admission:

- Validate a ticket at the door once. A second scan of the same ticket is rejected.

Out of scope for this version: transfers, resale, waitlists after sellout, and assigned seating changes after purchase.

## Non-functional requirements

- **No double sell.** At most one confirmed order owns a given seat for an event.
- **Hold window.** A hold lasts a fixed few minutes (target: 5–10). When it expires, the seat is purchasable again unless payment already succeeded.
- **Onsale load.** A single popular event can take a large burst of concurrent buyers. The waiting room absorbs arrivals the checkout path cannot serve yet.
- **Fair entry.** Queue position follows arrival order. A buyer who is admitted keeps that place until they finish or their session ends.
- **Seat map latency.** Rendering availability stays fast under read load. Availability shown on the map can lag the database by a short interval; checkout always checks the database (or the lock) before confirming a hold.
- **Durable orders.** A paid order survives process restarts. Seat state and order state can be reconstructed after a crash.
- **Idempotent checkout.** Retrying a payment or confirm request does not create a second order or charge.
- **Availability of browse.** Searching and viewing events keeps working when checkout is under heavy load.
