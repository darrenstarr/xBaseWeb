&& *  ============================================================
&& *  Customer Management
&& *  ============================================================

PROCEDURE CustomerMenu
  DO WHILE .T.
    CLEAR
    @ 1, 10 SAY "--- Customer Management ---"
    @ 3, 10 SAY "1. Add Customer"
    @ 4, 10 SAY "2. Find Customer"
    @ 5, 10 SAY "3. List All Customers"
    @ 6, 10 SAY "4. Customer Details"
    @ 7, 10 SAY "5. Risk Assessment"
    @ 9, 10 SAY "0. Back"
    @ 11,10 SAY "Select: " GET mChoice PICTURE "9"
    READ

    DO CASE
    CASE mChoice = 1
      DO AddCustomer
    CASE mChoice = 2
      DO FindCustomer
    CASE mChoice = 3
      DO ListCustomers
    CASE mChoice = 4
      DO ViewCustomer
    CASE mChoice = 5
      DO AssessRisk
    CASE mChoice = 0
      RETURN
    ENDCASE
  ENDDO
RETURN

PROCEDURE AddCustomer
  CLEAR
  @ 1, 10 SAY "--- New Customer ---"
  STORE "" TO mName
  STORE "" TO mAlias
  STORE "" TO mPhone
  STORE "" TO mEmail
  STORE 0 TO mRisk

  @ 4, 10 SAY "Name (required): " GET mName PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  @ 5, 10 SAY "Alias:           " GET mAlias PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXX"
  @ 6, 10 SAY "Phone:           " GET mPhone PICTURE "(XXX)XXX-XXXX"
  @ 7, 10 SAY "Email:           " GET mEmail PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  @ 8, 10 SAY "Initial Risk (0-100): " GET mRisk PICTURE "999"
  READ

  IF EMPTY(mName)
    WAIT "Name is required! Press any key..."
    RETURN
  ENDIF

  SELECT 0
  USE customers ALIAS cust
  APPEND BLANK
  REPLACE cust->name WITH mName
  REPLACE cust->alias WITH mAlias
  REPLACE cust->phone WITH mPhone
  REPLACE cust->email WITH mEmail
  REPLACE cust->risk_score WITH mRisk
  CLOSE DATABASES

  WAIT "Customer added! Press any key..."
RETURN

PROCEDURE FindCustomer
  CLEAR
  STORE "" TO mSearch
  @ 2, 10 SAY "Search by name: " GET mSearch PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  READ

  IF EMPTY(mSearch)
    RETURN
  ENDIF

  SELECT 0
  USE customers ALIAS cust
  LOCATE FOR UPPER(cust->name) = UPPER(mSearch)

  IF .NOT. FOUND()
    @ 5, 10 SAY "No customer found with that name."
    WAIT
    CLOSE DATABASES
    RETURN
  ENDIF

  DO DisplayCustomer
  CLOSE DATABASES
RETURN

PROCEDURE ListCustomers
  CLEAR
  SELECT 0
  USE customers ALIAS cust
  @ 1, 5 SAY "ID  Name                           Phone            Risk"
  @ 2, 5 SAY "--- ------------------------------ ---------------- ----"
  STORE 3 TO mRow

  SCATTER TO mCust

  DO WHILE .NOT. EOF()
    IF mRow > 20
      WAIT "Press any key for more..."
      CLEAR
      @ 1, 5 SAY "ID  Name                           Phone            Risk"
      @ 2, 5 SAY "--- ------------------------------ ---------------- ----"
      STORE 3 TO mRow
    ENDIF
    @ mRow, 5 SAY cust->id PICTURE "999"
    @ mRow, 9 SAY cust->name PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    @ mRow, 40 SAY cust->phone PICTURE "(XXX)XXX-XXXX"
    @ mRow, 56 SAY cust->risk_score PICTURE "999"
    SKIP
    mRow = mRow + 1
  ENDDO

  IF mRow = 3
    @ 4, 10 SAY "No customers found."
  ENDIF

  @ mRow + 1, 10 SAY "Press any key..."
  WAIT ""
  CLOSE DATABASES
RETURN

PROCEDURE ViewCustomer
  STORE 0 TO mId
  @ 2, 10 SAY "Customer ID: " GET mId PICTURE "9999"
  READ

  IF mId = 0
    RETURN
  ENDIF

  SELECT 0
  USE customers ALIAS cust
  GO mId

  IF EOF()
    @ 4, 10 SAY "Customer not found."
    WAIT
    CLOSE DATABASES
    RETURN
  ENDIF

  DO DisplayCustomer
  CLOSE DATABASES
RETURN

PROCEDURE DisplayCustomer
  @ 4, 10 SAY "ID:     " + STR(cust->id)
  @ 5, 10 SAY "Name:   " + cust->name
  @ 6, 10 SAY "Alias:  " + cust->alias
  @ 7, 10 SAY "Phone:  " + cust->phone
  @ 8, 10 SAY "Email:  " + cust->email
  @ 9, 10 SAY "Risk:   " + STR(cust->risk_score)
  @ 10,10 SAY "Notes:  " + cust->notes

  && Show their appointment history
  @ 12,10 SAY "--- Recent Appointments ---"
  STORE 13 TO mRow

  SELECT 0
  USE appointments ALIAS apt
  LOCATE FOR apt->customer_id = cust->id

  DO WHILE FOUND()
    IF mRow > 20
      WAIT "Press any key for more..."
      STORE 13 TO mRow
      @ 12,10 SAY "--- Recent Appointments ---"
    ENDIF
    @ mRow, 10 SAY apt->scheduled_for + " - " + apt->status
    mRow = mRow + 1
    CONTINUE
  ENDDO
  CLOSE DATABASES
RETURN

PROCEDURE AssessRisk
  && Calculate risk score based on various factors
  SELECT 0
  USE customers ALIAS cust
  USE appointments ALIAS apt IN 0
  USE invoices ALIAS inv IN 0

  DO WHILE .NOT. EOF()
    && Base risk from customer record
    STORE cust->risk_score TO mRisk

    && Increase risk for no-shows
    SELECT apt
    COUNT FOR apt->customer_id = cust->id .AND. apt->status = "no_show" TO mNoShows
    mRisk = mRisk + (mNoShows * 10)

    && Increase risk for overdue balances
    SELECT inv
    SUM inv->balance FOR inv->customer_id = cust->id .AND. inv->status = "overdue" TO mOverdue
    IF mOverdue > 0
      mRisk = mRisk + 15
    ENDIF

    && Cap at 100
    IF mRisk > 100
      mRisk = 100
    ENDIF

    IF mRisk <> cust->risk_score
      SELECT cust
      REPLACE cust->risk_score WITH mRisk
    ENDIF

    SELECT cust
    SKIP
  ENDDO

  CLOSE DATABASES
  WAIT "Risk assessment complete. Press any key..."
RETURN
