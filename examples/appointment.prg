&& *  ============================================================
&& *  Appointment Scheduling
&& *  ============================================================

PROCEDURE ScheduleAppointment
  CLEAR
  @ 1, 10 SAY "--- Schedule Appointment ---"

  && Pick a customer
  STORE 0 TO mCustId
  @ 3, 10 SAY "Customer ID: " GET mCustId PICTURE "9999"
  READ

  IF mCustId = 0
    RETURN
  ENDIF

  SELECT 0
  USE customers ALIAS cust
  GO mCustId
  IF EOF()
    @ 5, 10 SAY "Customer not found!"
    WAIT
    CLOSE DATABASES
    RETURN
  ENDIF
  @ 5, 10 SAY "Customer: " + cust->name

  && Check risk score — flag high-risk customers
  IF cust->risk_score >= 70
    @ 6, 10 SAY "WARNING: This customer has a high risk score of " + STR(cust->risk_score)
    @ 7, 10 SAY "Consider pre-payment requirement."
    WAIT "Press any key to continue..."
  ENDIF
  CLOSE DATABASES

  && Pick a service
  DO SelectService

  && Confirm
  STORE "N" TO mConfirm
  @ 16, 10 SAY "Schedule this appointment? (Y/N): " GET mConfirm PICTURE "!"
  READ

  IF mConfirm = "Y"
    SELECT 0
    USE appointments ALIAS apt
    APPEND BLANK
    REPLACE apt->customer_id WITH mCustId
    REPLACE apt->service_id WITH mSvcId
    REPLACE apt->scheduled_for WITH mDateTime
    REPLACE apt->duration WITH mDuration
    REPLACE apt->status WITH "pending"
    CLOSE DATABASES

    && Generate a waiver if required
    SELECT 0
    USE services ALIAS svc
    GO mSvcId
    IF svc->requires_waiver = 1
      SELECT 0
      USE appointments ALIAS apt
      GO BOTTOM
      REPLACE apt->waiver_signed WITH 0
      @ 18, 10 SAY "Waiver required! Please have customer sign."
    ENDIF
    CLOSE DATABASES

    WAIT "Appointment scheduled! Press any key..."
  ENDIF
RETURN

PROCEDURE SelectService
  CLEAR
  @ 1, 10 SAY "--- Available Services ---"
  @ 2, 10 SAY "ID  Service                    Duration  Price     Intensity"
  @ 3, 10 SAY "--- -------------------------  --------  --------  ---------"

  SELECT 0
  USE services ALIAS svc
  STORE 4 TO mRow

  DO WHILE .NOT. EOF()
    @ mRow, 2 SAY STR(svc->id)
    @ mRow, 6 SAY svc->name
    @ mRow, 32 SAY STR(svc->duration) + " min"
    @ mRow, 42 SAY svc->base_price PICTURE "9999.99"
    @ mRow, 52 SAY svc->intensity PICTURE "99"
    SKIP
    mRow = mRow + 1
  ENDDO

  && Get date/time
  STORE 0 TO mSvcId
  STORE "" TO mDateTime
  STORE 0 TO mDuration

  @ mRow + 1, 10 SAY "Service ID: " GET mSvcId PICTURE "99"
  READ

  IF mSvcId = 0
    CLOSE DATABASES
    RETURN
  ENDIF

  GO mSvcId
  IF EOF()
    @ mRow + 2, 10 SAY "Service not found!"
    WAIT
    CLOSE DATABASES
    RETURN
  ENDIF

  @ mRow + 2, 10 SAY "Selected: " + svc->name
  STORE svc->duration TO mDuration

  @ mRow + 3, 10 SAY "Date/Time (YYYY-MM-DD HH:MM): " GET mDateTime PICTURE "9999-99-99 99:99"
  READ

  CLOSE DATABASES
RETURN

PROCEDURE ViewAppointments
  CLEAR
  @ 1, 10 SAY "--- Upcoming Appointments ---"
  @ 2, 10 SAY "Date/Time            Customer     Service              Status"
  @ 3, 10 SAY "--------------------  -----------  -------------------  --------"

  SELECT 0
  USE appointments ALIAS apt
  USE customers ALIAS cust IN 0
  USE services ALIAS svc IN 0

  && Only show today and future
  LOCATE FOR apt->scheduled_for >= DATE()
  STORE 4 TO mRow

  DO WHILE FOUND() .AND. mRow < 22
    SELECT cust
    GO apt->customer_id

    SELECT svc
    GO apt->service_id

    @ mRow, 2 SAY apt->scheduled_for
    @ mRow, 21 SAY cust->name
    @ mRow, 36 SAY svc->name
    @ mRow, 56 SAY apt->status

    SELECT apt
    CONTINUE
    mRow = mRow + 1
  ENDDO

  IF mRow = 4
    @ 5, 10 SAY "No upcoming appointments."
  ENDIF

  CLOSE DATABASES
  WAIT ""
RETURN

PROCEDURE ServiceCatalog
  CLEAR
  @ 1, 10 SAY "--- Service Catalog ---"
  SELECT 0
  USE services ALIAS svc

  DO WHILE .NOT. EOF()
    CLEAR
    @ 2, 10 SAY "Name:        " + svc->name
    @ 3, 10 SAY "Description: " + svc->description
    @ 4, 10 SAY "Duration:    " + STR(svc->duration) + " minutes"
    @ 5, 10 SAY "Price:       $" + STR(svc->base_price)
    @ 6, 10 SAY "Intensity:   " + STR(svc->intensity) + "/10"
    @ 7, 10 SAY "Waiver:      " + IIF(svc->requires_waiver = 1, "Required", "Not required")
    @ 9, 10 SAY "Press any key for next service, Esc to exit..."
    WAIT ""
    SKIP
  ENDDO

  CLOSE DATABASES
RETURN
