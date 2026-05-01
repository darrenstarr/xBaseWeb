&& *  ============================================================
&& *  Invoicing & Collections
&& *  ============================================================

PROCEDURE GenerateInvoice
  CLEAR
  @ 1, 10 SAY "--- Generate Invoice ---"

  STORE 0 TO mAptId
  @ 3, 10 SAY "Appointment ID: " GET mAptId PICTURE "9999"
  READ

  IF mAptId = 0
    RETURN
  ENDIF

  SELECT 0
  USE appointments ALIAS apt
  GO mAptId
  IF EOF()
    @ 5, 10 SAY "Appointment not found!"
    WAIT
    CLOSE DATABASES
    RETURN
  ENDIF

  IF apt->status <> "completed"
    @ 5, 10 SAY "Appointment is not completed (status: " + apt->status + ")"
    @ 6, 10 SAY "Complete the session before generating an invoice."
    WAIT
    CLOSE DATABASES
    RETURN
  ENDIF

  && Lookup service for pricing
  SELECT 0
  USE services ALIAS svc
  GO apt->service_id

  && Calculate amounts
  STORE svc->base_price TO mSubtotal
  STORE mSubtotal * 0.08 TO mTax    * Florida sales tax, naturally
  STORE mSubtotal + mTax TO mTotal

  && Create invoice
  SELECT 0
  USE invoices ALIAS inv
  APPEND BLANK
  REPLACE inv->customer_id WITH apt->customer_id
  REPLACE inv->appointment_id WITH mAptId
  REPLACE inv->invoice_date WITH date()
  REPLACE inv->due_date WITH date() + 30
  REPLACE inv->subtotal WITH mSubtotal
  REPLACE inv->tax WITH mTax
  REPLACE inv->total WITH mTotal
  REPLACE inv->paid WITH 0
  REPLACE inv->balance WITH mTotal
  REPLACE inv->status WITH "pending"

  && Add line item
  SELECT 0
  USE invoice_items ALIAS itm
  APPEND BLANK
  REPLACE itm->invoice_id WITH inv->id
  REPLACE itm->description WITH svc->name + " — " + svc->description
  REPLACE itm->quantity WITH 1
  REPLACE itm->unit_price WITH mSubtotal
  REPLACE itm->total WITH mSubtotal

  CLOSE DATABASES

  @ 8, 10 SAY "Invoice generated!"
  @ 9, 10 SAY "Total: $" + STR(mTotal)
  @ 10,10 SAY "Due: " + DTOC(date() + 30)
  WAIT
RETURN

PROCEDURE CollectionsDashboard
  CLEAR
  @ 1, 10 SAY "=== Collections Dashboard ==="

  SELECT 0
  USE invoices ALIAS inv
  USE customers ALIAS cust IN 0

  && Count overdue invoices
  COUNT FOR inv->status = "overdue" TO mOverdueCount
  SUM inv->balance FOR inv->status = "overdue" TO mOverdueTotal

  && Count accounts in collections
  COUNT FOR inv->dunning_level >= 3 TO mCollectionsCount

  @ 3, 10 SAY "Overdue Invoices:     " + STR(mOverdueCount)
  @ 4, 10 SAY "Total Overdue Amount: $" + STR(mOverdueTotal)
  @ 5, 10 SAY "Accounts in Collections: " + STR(mCollectionsCount)

  @ 7, 10 SAY "1. Send Reminders"
  @ 8, 10 SAY "2. Escalate Dunning"
  @ 9, 10 SAY "3. Record Payment"
  @ 10,10 SAY "4. Collection Actions Log"
  @ 11,10 SAY "5. Run Collections Report"
  @ 13,10 SAY "0. Back"

  STORE 0 TO mChoice
  @ 15,10 SAY "Select: " GET mChoice PICTURE "9"
  READ

  DO CASE
  CASE mChoice = 1
    DO SendReminders
  CASE mChoice = 2
    DO EscalateDunning
  CASE mChoice = 3
    DO RecordPayment
  CASE mChoice = 4
    DO CollectionLog
  CASE mChoice = 5
    DO CollectionsReport
  ENDCASE

  CLOSE DATABASES
RETURN

PROCEDURE OverdueAccounts
  CLEAR
  @ 1, 10 SAY "=== Overdue Accounts ==="
  @ 2, 10 SAY "Customer                Invoice   Amount    Overdue   Dunning"
  @ 3, 10 SAY "----------------------  --------  --------  --------  -------"

  SELECT 0
  USE customers ALIAS cust
  USE invoices ALIAS inv IN 0

  && Simple scan — not using the VIEW for now
  SELECT inv
  LOCATE FOR inv->balance > 0 .AND. inv->due_date < date()

  STORE 4 TO mRow
  DO WHILE FOUND() .AND. mRow < 22
    SELECT cust
    GO inv->customer_id

    @ mRow, 2 SAY cust->name
    SELECT inv
    @ mRow, 25 SAY STR(inv->id)
    @ mRow, 34 SAY inv->balance PICTURE "9999.99"
    @ mRow, 44 SAY STR(INT(date() - inv->due_date)) + " days"
    @ mRow, 54 SAY inv->dunning_level PICTURE "99"

    CONTINUE
    mRow = mRow + 1
  ENDDO

  IF mRow = 4
    @ 5, 10 SAY "No overdue accounts. Amazing."
  ENDIF

  CLOSE DATABASES
  WAIT ""
RETURN

PROCEDURE SendReminders
  CLEAR
  @ 1, 10 SAY "--- Sending Payment Reminders ---"

  SELECT 0
  USE invoices ALIAS inv
  USE customers ALIAS cust IN 0

  LOCATE FOR inv->status = "overdue" .AND. inv->dunning_level = 0
  STORE 0 TO mSent

  DO WHILE FOUND()
    SELECT cust
    GO inv->customer_id

    @ 3, 10 SAY "Reminding: " + cust->name

    && Log the collection action
    SELECT 0
    USE collection_actions ALIAS ca
    APPEND BLANK
    REPLACE ca->invoice_id WITH inv->id
    REPLACE ca->action_type WITH "email"
    REPLACE ca->contact WITH cust->email
    REPLACE ca->response WITH ""
    REPLACE ca->result WITH "sent"
    CLOSE DATABASES

    SELECT 0
    USE invoices ALIAS inv
    GO inv->id
    REPLACE inv->dunning_level WITH 1
    mSent = mSent + 1

    CONTINUE
  ENDDO

  @ 5, 10 SAY "Sent " + STR(mSent) + " reminders."
  CLOSE DATABASES
  WAIT ""
RETURN

PROCEDURE EscalateDunning
  CLEAR
  @ 1, 10 SAY "--- Escalate Dunning Level ---"

  SELECT 0
  USE invoices ALIAS inv
  USE customers ALIAS cust IN 0

  && Find invoices past due that need escalation
  && Level 0 -> 1 after 15 days overdue
  && Level 1 -> 2 after 30 days overdue
  && Level 2 -> 3 after 60 days overdue (collections)
  && Level 3 -> 4 after 90 days (psychic message)

  STORE date() TO mToday
  LOCATE FOR inv->status <> "paid" .AND. inv->balance > 0

  DO WHILE FOUND()
    STORE mToday - inv->due_date TO mDaysOverdue
    STORE inv->dunning_level TO mLevel
    STORE 0 TO mChange

    IF mDaysOverdue >= 90 .AND. mLevel < 4
      REPLACE inv->dunning_level WITH 4
      mChange = 1
    ELSE
      IF mDaysOverdue >= 60 .AND. mLevel < 3
        REPLACE inv->dunning_level WITH 3
        mChange = 1
      ELSE
        IF mDaysOverdue >= 30 .AND. mLevel < 2
          REPLACE inv->dunning_level WITH 2
          mChange = 1
        ELSE
          IF mDaysOverdue >= 15 .AND. mLevel < 1
            REPLACE inv->dunning_level WITH 1
            mChange = 1
          ENDIF
        ENDIF
      ENDIF
    ENDIF

    IF mChange = 1
      SELECT cust
      GO inv->customer_id
      @ 3, 10 SAY "Escalated: " + cust->name + " to Level " + STR(inv->dunning_level)
    ENDIF

    SELECT inv
    CONTINUE
  ENDDO

  CLOSE DATABASES
  WAIT "Escalation complete. Press any key..."
RETURN

PROCEDURE RecordPayment
  CLEAR
  @ 1, 10 SAY "--- Record Payment ---"

  STORE 0 TO mInvId
  @ 3, 10 SAY "Invoice ID: " GET mInvId PICTURE "9999"
  READ

  IF mInvId = 0
    RETURN
  ENDIF

  SELECT 0
  USE invoices ALIAS inv
  GO mInvId
  IF EOF()
    @ 5, 10 SAY "Invoice not found!"
    WAIT
    CLOSE DATABASES
    RETURN
  ENDIF

  @ 5, 10 SAY "Invoice Total:  $" + STR(inv->total)
  @ 6, 10 SAY "Current Paid:   $" + STR(inv->paid)
  @ 7, 10 SAY "Balance:        $" + STR(inv->balance)

  STORE 0 TO mAmount
  STORE "cash" TO mMethod
  @ 9, 10 SAY "Payment Amount: $" GET mAmount PICTURE "9999.99"
  @ 10,10 SAY "Method (cash/card/crypto): " GET mMethod PICTURE "XXXXXXXXXXXXXXXX"
  READ

  IF mAmount <= 0
    RETURN
  ENDIF

  && Record payment
  SELECT 0
  USE payments ALIAS pmt
  APPEND BLANK
  REPLACE pmt->invoice_id WITH mInvId
  REPLACE pmt->amount WITH mAmount
  REPLACE pmt->method WITH mMethod
  CLOSE DATABASES

  && Update invoice
  SELECT inv
  REPLACE inv->paid WITH inv->paid + mAmount
  REPLACE inv->balance WITH inv->total - inv->paid

  IF inv->balance <= 0
    REPLACE inv->status WITH "paid"
    REPLACE inv->dunning_level WITH 0
    @ 12, 10 SAY "Invoice PAID IN FULL."
  ELSE
    IF inv->balance < inv->total
      REPLACE inv->status WITH "partial"
      @ 12, 10 SAY "Partial payment recorded. Remaining: $" + STR(inv->balance)
    ENDIF
  ENDIF

  CLOSE DATABASES
  WAIT ""
RETURN

PROCEDURE CollectionLog
  CLEAR
  @ 1, 10 SAY "--- Collection Actions Log ---"

  SELECT 0
  USE collection_actions ALIAS ca
  USE customers ALIAS cust IN 0
  USE invoices ALIAS inv IN 0

  STORE 3 TO mRow
  LOCATE FOR .T.

  DO WHILE FOUND() .AND. mRow < 22
    SELECT inv
    GO ca->invoice_id
    SELECT cust
    GO inv->customer_id

    @ mRow, 2 SAY ca->acted_at
    @ mRow, 16 SAY LEFT(cust->name, 16)
    @ mRow, 34 SAY ca->action_type
    @ mRow, 46 SAY ca->result

    SELECT ca
    CONTINUE
    mRow = mRow + 1
  ENDDO

  IF mRow = 3
    @ 4, 10 SAY "No collection actions recorded."
  ENDIF

  CLOSE DATABASES
  WAIT ""
RETURN

PROCEDURE CollectionsReport
  CLEAR
  @ 1, 10 SAY "=== Collections Report ==="

  SELECT 0
  USE invoices ALIAS inv
  USE customers ALIAS cust IN 0

  STORE 0 TO mTotalOverdue
  STORE 0 TO mInCollections
  STORE 0 TO mWrittenOff
  STORE 0 TO mCountOverdue

  LOCATE FOR inv->balance > 0
  DO WHILE FOUND()
    STORE mTotalOverdue + inv->balance TO mTotalOverdue
    STORE mCountOverdue + 1 TO mCountOverdue

    IF inv->dunning_level >= 3
      mInCollections = mInCollections + 1
    ENDIF

    IF inv->status = "written_off"
      mWrittenOff = mWrittenOff + 1
    ENDIF

    CONTINUE
  ENDDO

  @ 3, 10 SAY "Total Outstanding:     $" + STR(mTotalOverdue)
  @ 4, 10 SAY "Accounts Receivable:   " + STR(mCountOverdue)
  @ 5, 10 SAY "In Collections:        " + STR(mInCollections)
  @ 6, 10 SAY "Written Off:           " + STR(mWrittenOff)

  && Collection effectiveness
  STORE 0 TO mCollected
  SUM inv->paid FOR inv->status = "paid" TO mCollected

  @ 8, 10 SAY "Total Collected:       $" + STR(mCollected)

  IF mTotalOverdue + mCollected > 0
    STORE (mCollected / (mTotalOverdue + mCollected)) * 100 TO mRate
    @ 9, 10 SAY "Collection Rate:       " + STR(mRate) + "%"
  ENDIF

  CLOSE DATABASES
  WAIT ""
RETURN
