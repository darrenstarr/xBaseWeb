&& Application — all logic in xBase

PROCEDURE MainMenu
  SET TITLE TO "The DeSantis Cure for Woke"
  SET TAGLINE TO "We'll straighten you out."
  MENU "Main Menu" "Customer Management" -> CustomerMenu, "Appointments" -> ApptMenu, "Services" -> ServicesMenu, "Invoicing & Collections" -> InvoiceMenu
RETURN

PROCEDURE CustomerMenu
  MENU "Customers" "Add Customer" -> AddCustomer, "List Customers" -> ListCustomers
RETURN

PROCEDURE AddCustomer
  IF mName == ""
    CLEAR
    @ 1, 1 SAY "--- New Customer ---"
    @ 3, 1 SAY "Name:  " GET mName PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    @ 4, 1 SAY "Alias: " GET mAlias PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXX"
    @ 5, 1 SAY "Phone: " GET mPhone PICTURE "(XXX)XXX-XXXX"
    @ 6, 1 SAY "Email: " GET mEmail PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    READ
    RETURN
  ENDIF
  SELECT 0
  USE customers ALIAS cust
  APPEND BLANK
  REPLACE cust->name WITH mName
  REPLACE cust->alias WITH mAlias
  REPLACE cust->phone WITH mPhone
  REPLACE cust->email WITH mEmail
  CLOSE DATABASES
  CLEAR
  @ 2, 1 SAY "Customer saved!"
  WAIT ""
RETURN

PROCEDURE EditCustomer
  IF mName == "" .AND. mId == ""
    CLEAR
    @ 1, 1 SAY "Customer ID: " GET mId PICTURE "9999"
    READ
    RETURN
  ENDIF
  IF mEmail <> ""
    SELECT 0
    USE customers ALIAS cust
    GO VAL(mId)
    REPLACE cust->name WITH mName
    REPLACE cust->alias WITH mAlias
    REPLACE cust->phone WITH mPhone
    REPLACE cust->email WITH mEmail
    CLOSE DATABASES
    CLEAR
    @ 2, 1 SAY "Customer updated!"
    WAIT ""
    RETURN
  ENDIF
  && Show edit form with pre-filled values from row action
  CLEAR
  @ 1, 1 SAY "--- Edit Customer ---"
  @ 3, 1 SAY "Name:  " GET mName PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  @ 4, 1 SAY "Alias: " GET mAlias PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXX"
  @ 5, 1 SAY "Phone: " GET mPhone PICTURE "(XXX)XXX-XXXX"
  @ 6, 1 SAY "Email: " GET mEmail PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  READ
RETURN

PROCEDURE ListCustomers
  CLEAR
  @ 1, 1 SAY "--- Customer List ---"
  RUNSQL "SELECT id, name, alias, phone, risk_score FROM customers ORDER BY name, id" COLUMNS "ID", "Name", "Alias", "Phone", "Risk" SEARCH "Name", "Alias", "Phone", "Email" ACTIONS "Edit" -> "EditCustomer", "Delete" -> "DeleteCustomer"
RETURN

PROCEDURE DeleteCustomer
  IF mId == ""
    CLEAR
    @ 1, 1 SAY "Customer ID: " GET mId PICTURE "9999"
    READ
    RETURN
  ENDIF
  IF _confirm == ""
    CONFIRM "Delete this customer?"
    RETURN
  ENDIF
  IF _confirm == "yes"
    SELECT 0
    USE customers ALIAS cust
    GO VAL(mId)
    DELETE
    PACK
    CLOSE DATABASES
    CLEAR
    @ 2, 1 SAY "Customer deleted."
  ELSE
    CLEAR
    @ 2, 1 SAY "Delete cancelled."
  ENDIF
  WAIT ""
RETURN

PROCEDURE ApptMenu
  MENU "Appointments" "Schedule Appointment" -> AddAppointment, "List Appointments" -> ListAppointments
RETURN

PROCEDURE AddAppointment
  IF mCustId == ""
    CLEAR
    @ 1, 1 SAY "--- Schedule Appointment ---"
    @ 3, 1 SAY "Customer ID: " GET mCustId PICTURE "9999"
    @ 4, 1 SAY "Service ID:  " GET mSvcId PICTURE "99"
    @ 5, 1 SAY "Date/Time:   " GET mScheduled PICTURE "9999-99-99 99:99"
    READ
    RETURN
  ENDIF
  SELECT 0
  USE appointments ALIAS apt
  APPEND BLANK
  REPLACE apt->customer_id WITH VAL(mCustId)
  REPLACE apt->service_id WITH VAL(mSvcId)
  REPLACE apt->scheduled_for WITH mScheduled
  REPLACE apt->status WITH "pending"
  CLOSE DATABASES
  CLEAR
  @ 2, 1 SAY "Appointment scheduled!"
  WAIT ""
RETURN

PROCEDURE ListAppointments
  CLEAR
  @ 1, 1 SAY "--- Appointment List ---"
  RUNSQL "SELECT a.id, c.name, s.name, a.scheduled_for, a.status FROM appointments a JOIN customers c ON c.id=a.customer_id JOIN services s ON s.id=a.service_id ORDER BY a.scheduled_for DESC, a.id DESC" COLUMNS "ID", "Customer", "Service", "Scheduled", "Status" ACTIONS "Complete" -> "CompleteAppt", "Cancel" -> "CancelAppt", "Delete" -> "DeleteAppt"
RETURN

PROCEDURE ServicesMenu
  MENU "Services" "Add Service" -> AddService, "List Services" -> ListServices
RETURN

PROCEDURE AddService
  IF mName == ""
    CLEAR
    @ 1, 1 SAY "--- New Service ---"
    @ 3, 1 SAY "Name:        " GET mName PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXX"
    @ 4, 1 SAY "Description: " GET mDesc PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    @ 5, 1 SAY "Duration:    " GET mDuration PICTURE "999"
    @ 6, 1 SAY "Price:       " GET mPrice PICTURE "9999.99"
    @ 7, 1 SAY "Intensity:   " GET mIntensity PICTURE "99"
    READ
    RETURN
  ENDIF
  SELECT 0
  USE services ALIAS svc
  APPEND BLANK
  REPLACE svc->name WITH mName
  REPLACE svc->description WITH mDesc
  REPLACE svc->duration WITH VAL(mDuration)
  REPLACE svc->base_price WITH VAL(mPrice)
  REPLACE svc->intensity WITH VAL(mIntensity)
  CLOSE DATABASES
  CLEAR
  @ 2, 1 SAY "Service saved!"
  WAIT ""
RETURN

PROCEDURE EditService
  IF mName == "" .AND. mId == ""
    CLEAR
    @ 1, 1 SAY "Service ID: " GET mId PICTURE "99"
    READ
    RETURN
  ENDIF
  SELECT 0
  USE services ALIAS svc
  LOCATE FOR svc->id = VAL(mId)
  IF FOUND()
    CLEAR
    @ 1, 1 SAY "--- Edit Service ---"
    @ 3, 1 SAY "Name:        " GET svc->name PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXX"
    @ 4, 1 SAY "Description: " GET svc->desc PICTURE "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    @ 5, 1 SAY "Duration:    " GET svc->duration PICTURE "999"
    @ 6, 1 SAY "Price:       " GET svc->base_price PICTURE "9999.99"
    @ 7, 1 SAY "Intensity:   " GET svc->intensity PICTURE "99"
    READ
    IF svc->name <> ""
      REPLACE svc->name WITH svc->name
    ENDIF
    CLOSE DATABASES
    CLEAR
    @ 2, 1 SAY "Service updated!"
    WAIT ""
  ELSE
    CLEAR
    @ 2, 1 SAY "Service not found."
    WAIT ""
  ENDIF
RETURN

PROCEDURE ListServices
  CLEAR
  @ 1, 1 SAY "--- Service List ---"
  RUNSQL "SELECT id, name, description, base_price, intensity FROM services ORDER BY id" COLUMNS "ID", "Name", "Description", "Price", "Intensity" ACTIONS "Edit" -> "EditService", "Delete" -> "DeleteService"
RETURN

PROCEDURE DeleteService
  IF mId == ""
    CLEAR
    @ 1, 1 SAY "Service ID: " GET mId PICTURE "99"
    READ
    RETURN
  ENDIF
  SELECT 0
  USE services ALIAS svc
  LOCATE FOR svc->id = VAL(mId)
  IF FOUND()
    DELETE
    PACK
    CLEAR
    @ 2, 1 SAY "Service deleted."
  ELSE
    CLEAR
    @ 2, 1 SAY "Service not found."
  ENDIF
  CLOSE DATABASES
  WAIT ""
RETURN

PROCEDURE InvoiceMenu
  MENU "Invoicing" "List Invoices" -> ListInvoices, "Overdue Accounts" -> OverdueAccounts, "Generate Invoice" -> GenerateInvoice
RETURN

PROCEDURE ListInvoices
  CLEAR
  @ 1, 1 SAY "--- Invoice List ---"
  RUNSQL "SELECT i.id, c.name, i.total, i.paid, i.balance, i.status, i.dunning_level, i.due_date FROM invoices i JOIN customers c ON c.id=i.customer_id ORDER BY i.due_date DESC" COLUMNS "#", "Customer", "Total", "Paid", "Balance", "Status", "Dunning", "Due" ACTIONS "Pay" -> "RecordPayment", "Delete" -> "DeleteInvoice"
RETURN

PROCEDURE OverdueAccounts
  CLEAR
  @ 1, 1 SAY "--- Overdue Accounts ---"
  RUNSQL "SELECT invoice_id, customer_name, phone, balance, days_overdue FROM overdue_invoices ORDER BY days_overdue DESC" COLUMNS "Inv #", "Customer", "Phone", "Balance", "Days Overdue"
RETURN

PROCEDURE GenerateInvoice
  IF mAptId == ""
    CLEAR
    @ 1, 1 SAY "Completed Appointment ID: " GET mAptId PICTURE "9999"
    READ
    RETURN
  ENDIF
  SELECT 0
  USE appointments ALIAS apt
  LOCATE FOR apt->id = VAL(mAptId)
  IF FOUND() .AND. apt->status = "completed"
    SELECT 0
    USE services ALIAS svc
    LOCATE FOR svc->id = apt->service_id
    IF FOUND()
      SELECT 0
      USE invoices ALIAS inv
      APPEND BLANK
      REPLACE inv->customer_id WITH apt->customer_id
      REPLACE inv->appointment_id WITH VAL(mAptId)
      REPLACE inv->total WITH svc->base_price * 1.08
      REPLACE inv->balance WITH svc->base_price * 1.08
      REPLACE inv->status WITH "pending"
      CLOSE DATABASES
      CLEAR
      @ 2, 1 SAY "Invoice generated!"
    ELSE
      CLOSE DATABASES
      CLEAR
      @ 2, 1 SAY "Service not found."
    ENDIF
  ELSE
    CLOSE DATABASES
    CLEAR
    @ 2, 1 SAY "Appointment not found or not completed."
  ENDIF
  WAIT ""
RETURN
