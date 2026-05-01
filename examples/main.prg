&& *  ============================================================
&& *  Ron DeSantis's Cure for Woke — Main Program
&& *  A scheduling & invoicing system for a torture dungeon.
&& *  ============================================================

PROCEDURE Main
  SET TALK OFF
  SET SAFETY OFF
  SET DELETED ON

  DO InitDatabase
  DO MainMenu
RETURN

PROCEDURE InitDatabase
  && Schema is applied by the runtime on first connection.
  && Tables: customers, services, appointments, invoices,
  &&         invoice_items, payments, collection_actions
  CLOSE DATABASES
  SELECT 0
  USE services ALIAS svc
  IF EOF()
    && Seed data on first run — services table
    APPEND BLANK
    REPLACE svc->name WITH "Waterboarding"
    REPLACE svc->base_price WITH 299.99
    APPEND BLANK
    REPLACE svc->name WITH "Iron Maiden"
    REPLACE svc->base_price WITH 499.99
    APPEND BLANK
    REPLACE svc->name WITH "Dolly Parton Rock"
    REPLACE svc->base_price WITH 399.99
    APPEND BLANK
    REPLACE svc->name WITH "Drawing & Quartering"
    REPLACE svc->base_price WITH 799.99
  ENDIF
  CLOSE DATABASES
RETURN

PROCEDURE MainMenu
  DO WHILE .T.
    CLEAR
    @ 1, 10 SAY "========================================"
    @ 2, 10 SAY "  Ron DeSantis's Cure for Woke v1.0"
    @ 3, 10 SAY "  'We'll straighten you out.'"
    @ 4, 10 SAY "========================================"
    @ 6, 10 SAY "1. Customer Management"
    @ 7, 10 SAY "2. Schedule Appointment"
    @ 8, 10 SAY "3. View Appointments"
    @ 9, 10 SAY "4. Generate Invoice"
    @ 10,10 SAY "5. Collections Dashboard"
    @ 11,10 SAY "6. Overdue Accounts"
    @ 12,10 SAY "7. Service Catalog"
    @ 14,10 SAY "0. Exit"
    @ 16,10 SAY "Select: " GET mChoice PICTURE "9"
    READ

    DO CASE
    CASE mChoice = 1
      DO CustomerMenu
    CASE mChoice = 2
      DO ScheduleAppointment
    CASE mChoice = 3
      DO ViewAppointments
    CASE mChoice = 4
      DO GenerateInvoice
    CASE mChoice = 5
      DO CollectionsDashboard
    CASE mChoice = 6
      DO OverdueAccounts
    CASE mChoice = 7
      DO ServiceCatalog
    CASE mChoice = 0
      QUIT
    OTHERWISE
      WAIT "Invalid selection. Press any key..."
    ENDCASE
  ENDDO
RETURN
